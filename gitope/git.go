package gitope

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/aak1247/gchangelog/configs"
	"github.com/aak1247/gchangelog/utils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

type Ref struct {
	Hash string
	When time.Time
}

type ChangeLog struct {
	Version string
	Head    *Ref
	Groups  map[string][]*object.Commit
}

func (c *ChangeLog) ParseCommits(commits []*object.Commit) {
	// 分组
	for _, commit := range commits {
		t := ParseCommitMessageType(commit)
		if !configs.MR {
			if strings.Contains(commit.Message, "Merge") {
				continue
			}
		}
		if g, ok := c.Groups[t]; ok == true {
			g = append(g, commit)
			c.Groups[t] = g
		} else {
			c.Groups[t] = []*object.Commit{commit}
		}
	}
}

// String 输出changelog
func (c *ChangeLog) String() string {
	s := strings.Builder{}
	s.WriteString(c.RenderVersionHeader())
	// 分类型输出
	for _, k := range configs.Types {
		if v, ok := c.Groups[k]; ok {
			if len(v) == 0 {
				continue
			}
			s.WriteString(fmt.Sprintf("### %s\n", k))
			// 用于去重
			msgMap := make(map[string]string)
			for _, commit := range v {
				msgMap[commit.Message] = commit.Hash.String()
			}
			// 拼接单条commit msg
			for _, v := range v {
				// 去重, 旧提交不输出
				if hash, ok := msgMap[v.Message]; ok {
					if hash != v.Hash.String() {
						continue
					}
				}

				commitMsg := c.RenderCommit(v)
				s.WriteString(commitMsg)
			}
		}
	}
	return s.String()
}

func (c *ChangeLog) RenderVersionHeader() string {
	return fmt.Sprintf("## %s    <sub>[%s](%s) - [%s](%s) %s</sub>\n\n", c.Version,
		c.Head.When.Format("2006-01-02"), GetTagUrl(configs.BaseUrl, configs.Project, c.Version),
		c.Head.Hash[:8], GetCommitUrl(configs.BaseUrl, configs.Project, c.Version),
		RenderPipelineUrl(configs.BaseUrl, configs.Project, c.Version))
}

func (c *ChangeLog) RenderCommit(v *object.Commit) string {
	s := strings.Builder{}
	msg := v.Message
	contents := make([]string, 0)
	hasContent := false
	// 多行处理
	if utils.IsMultiline(msg) {
		contents = strings.Split(msg, "\n")
		msg = contents[0]
		contents = contents[1:]
		for _, c := range contents {
			if strings.TrimSpace(c) != "" {
				hasContent = true
				break
			}
		}
	}
	// 输出
	s.WriteString(fmt.Sprintf("- %s ( [%s by %s](%s) ) - <sub>%s</sub>\n", msg, v.Hash.String()[:8], v.Author.Name, GetCommitUrl(configs.BaseUrl, configs.Project, v.Hash.String()), v.Author.When.Format("2006-01-02 15:04")))
	// 多行内容输出
	if hasContent {
		s.WriteString("  ```markdown\n")
		for _, v := range contents {
			s.WriteString(fmt.Sprintf("  %s\n", v))
		}
		s.WriteString("  ```\n")
	}
	return s.String()
}

func FindCommits(tag2 *plumbing.Reference, tag1 *plumbing.Reference, r *git.Repository) []*object.Commit {
	var tag1Hash, tag2Hash string
	var options = &git.LogOptions{
		From:  tag1.Hash(),
		Order: git.LogOrderDFS,
	}
	tag1Hash = tag1.Hash().String()
	var startTime, endTime time.Time
	var tag1Head, tag2Head *object.Tag
	var commitHead1, commitHead2 *object.Commit
	tag1Head, err := r.TagObject(tag1.Hash())
	if err != nil {
		if err == plumbing.ErrObjectNotFound {
			commitHead1, err = r.CommitObject(tag1.Hash())
			if err != nil {
				log.Println("failed to find tag ref ", err)
				endTime = time.Now()
			}
		}
		endTime = commitHead1.Author.When
	} else {
		endTime = tag1Head.Tagger.When
	}
	options.Until = &(endTime)
	if tag2 != nil {
		// 有旧tag
		tag2Hash = tag2.Hash().String()
		tag2Head, err = r.TagObject(tag2.Hash())
		if err != nil {
			if err == plumbing.ErrObjectNotFound {
				commitHead2, err = r.CommitObject(tag2.Hash())
				if err != nil {
					log.Println("failed to find tag ref ", err)
					startTime = time.UnixMilli(0)
				}
			}
			startTime = commitHead2.Author.When
		} else {
			startTime = tag2Head.Tagger.When
		}
		options.Since = &(startTime)
	}
	// 遍历两个tag中间的log, 通过hash
	logIter, err := r.Log(options)
	if err != nil {
		return make([]*object.Commit, 0)
	}
	commits := make([]*object.Commit, 0)
	var start, end bool
	for {
		commit, err := logIter.Next()
		if err != nil || commit == nil {
			break
		}
		if end {
			// do not print this
			// log.Println("branch ended")
		}
		if commit.Hash.String() == tag1Hash {
			// 开始
			start = true
		}
		if commit.Hash.String() == tag2Hash {
			// 结束
			end = true
		}
		if configs.SkipMsgs.ShouldSkip(commit.Message) {
			continue
		}
		if start {
			commits = append(commits, commit)
		}
	}
	return commits
}

func FindTag(err error, r *git.Repository) (*plumbing.Reference, *plumbing.Reference, error) {
	// 先拿到最近的两个tag
	var tag1, tag2 *plumbing.Reference
	var tag1Name string
	tagIter, err := r.Tags()
	if err != nil {
		panic(err)
	}
	tag1, err = tagIter.Next()
	tag1Name = TagName(tag1)
	// 遍历找到最后两个
	for {
		tagN, err := tagIter.Next()
		if err != nil || tagN == nil {
			break
		}
		tagNName := TagName(tagN)
		if VersionCompare(tagNName, tag1Name) > 0 {
			tag1Name = TagName(tag1)
			tag2 = tag1
			tag1 = tagN
		}
	}

	if tag1 == nil && tag2 == nil {
		// 没有tag
		panic("no tag found")
	}
	if err != nil {
		// 报错
		panic(err)
	}
	return tag1, tag2, err
}

func FindPreviousTag(r *git.Repository, currentTag *plumbing.Reference) (*plumbing.Reference, error) {
	// 先拿到最近的两个tag
	var tag2 *plumbing.Reference
	var tag2Name = "0.0.0"
	var currentTagName = TagName(currentTag)
	tagIter, err := r.Tags()
	if err != nil {
		panic(err)
	}
	// 遍历找到最后两个
	for {
		tagN, err := tagIter.Next()
		if err != nil || tagN == nil {
			break
		}
		tagNName := TagName(tagN)
		if VersionCompare(tagNName, tag2Name) > 0 && VersionCompare(tagNName, currentTagName) < 0 {
			tag2 = tagN
			tag2Name = tagNName
		}
	}

	if tag2 == nil {
		// 没有tag
		return nil, err
	}
	if err != nil {
		// 报错
		panic(err)
	}
	return tag2, err
}

func ParseCommitMessageType(commit *object.Commit) (typ string) {
	fullMsg := commit.Message
	// 根据configs.Types 解析类型：大小写不敏感，且必须紧跟冒号（不允许空格）
	for _, t := range configs.Types {
		if len(fullMsg) >= len(t)+1 {
			prefix := fullMsg[:len(t)]
			if strings.EqualFold(prefix, t) && fullMsg[len(t)] == ':' {
				return t
			}
		}
	}
	return "other"
}

func TagName(ref *plumbing.Reference) string {
	return strings.TrimPrefix(ref.Name().String(), "refs/tags/")
}

func GetProjectPath(r *git.Repository) string {
	remotes, err := r.Remotes()
	if err != nil {
		panic(err)
	}
	var fullUrl string
	//var remoteUrl string
	for _, remote := range remotes {
		if remote.Config().Name == "origin" {
			fullUrl = remote.Config().URLs[0]
		}
	}
	endpoint, err := transport.NewEndpoint(fullUrl)
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSuffix(strings.TrimPrefix(endpoint.Path, "/"), ".git")
}

func GetBaseUrl(r *git.Repository) string {
	remotes, err := r.Remotes()
	if err != nil {
		panic(err)
	}
	var fullUrl string
	//var remoteUrl string
	for _, remote := range remotes {
		if remote.Config().Name == "origin" {
			fullUrl = remote.Config().URLs[0]
		}
	}
	endpoint, err := transport.NewEndpoint(fullUrl)
	if err != nil {
		log.Fatal(err)
	}
	if endpoint.Protocol == "http" {
		configs.HTTP = true
	}
	baseUrl := endpoint.Host
	if configs.HTTP {
		baseUrl = "http://" + baseUrl
	} else {
		baseUrl = "https://" + baseUrl
	}
	if endpoint.Port != 0 && strings.Contains(endpoint.Protocol, "http") {
		baseUrl += ":" + strconv.Itoa(endpoint.Port)
	}
	return baseUrl
}

func GetCommitUrl(base, project, hash string) string {
	if strings.Contains(base, "gitlab") {
		return fmt.Sprintf("%s/%s/-/commits/%s", base, project, hash)
	}
	if strings.Contains(base, "github") {
		return fmt.Sprintf("%s/%s/commit/%s", base, project, hash)
	}
	return fmt.Sprintf("%s/%s/commits/%s", base, project, hash)
}

func GetTagUrl(base, project, tagName string) string {
	if strings.Contains(base, "gitlab") {
		return fmt.Sprintf("%s/%s/-/tags/%s", base, project, tagName)
	}
	if strings.Contains(base, "github") {
		return fmt.Sprintf("%s/%s/releases/tag/%s", base, project, tagName)
	}

	return fmt.Sprintf("%s/%s/-/tags/%s", base, project, tagName)
}

func GetTagPipelineUrl(base, project, tagName string) string {
	return fmt.Sprintf("%s/%s/pipelines?page=1&scope=tags&ref=%s", base, project, tagName)
}

func RenderPipelineUrl(base, project, tagName string) string {
	if strings.Contains(base, "gitlab") {
		return fmt.Sprintf("[![CI](%s/%s/badges/%s/pipeline.svg?ignore_skipped=true)](%s)",
			base, project, tagName, GetTagPipelineUrl(base, project, tagName))
	} else if strings.Contains(base, "github") {
		return fmt.Sprintf("[![GH Action](%s/%s/main.yml/badge.svg?branch=%s)](%s/%s/actions)", base, project, tagName, base, project)
	}
	return "unknown"
}

// VersionCompare 版本大于
func VersionCompare(v1, v2 string) int {
	normalize := func(s string) string {
		// 去掉 v/V 前缀
		s = strings.TrimPrefix(s, "v")
		s = strings.TrimPrefix(s, "V")
		// 去掉产品名前缀（直到第一个数字）
		if idx := strings.IndexFunc(s, func(r rune) bool { return unicode.IsDigit(r) }); idx > 0 {
			s = s[idx:]
		}
		// 统一分隔符：下划线、加号视为点
		s = strings.ReplaceAll(s, "_", ".")
		s = strings.ReplaceAll(s, "+", ".")
		return s
	}

	v1 = normalize(v1)
	v2 = normalize(v2)

	// 如果完全相同，直接返回0
	if v1 == v2 {
		return 0
	}

	// 以点或短横分割所有token（短横用于先行版本）
	sep := regexp.MustCompile(`[\.-]`)
	t1 := sep.Split(v1, -1)
	t2 := sep.Split(v2, -1)

	// 特殊处理：比较前三个数字版本（major.minor.patch）
	var v1Nums, v2Nums []int
	for i := 0; i < 3 && i < len(t1); i++ {
		if num, err := strconv.Atoi(t1[i]); err == nil {
			v1Nums = append(v1Nums, num)
		} else {
			break
		}
	}
	for i := 0; i < 3 && i < len(t2); i++ {
		if num, err := strconv.Atoi(t2[i]); err == nil {
			v2Nums = append(v2Nums, num)
		} else {
			break
		}
	}

	// 比较数字版本部分
	for i := 0; i < len(v1Nums) && i < len(v2Nums); i++ {
		if v1Nums[i] < v2Nums[i] {
			return -1
		} else if v1Nums[i] > v2Nums[i] {
			return 1
		}
	}

	// 数字版本相同，比较长度
	if len(v1Nums) < len(v2Nums) {
		return -1
	} else if len(v1Nums) > len(v2Nums) {
		return 1
	}

	// 数字版本完全相同，检查是否有先行版本标识
	var v1HasPre, v2HasPre bool
	for i := 3; i < len(t1); i++ {
		if !isNumeric(t1[i]) {
			v1HasPre = true
			break
		}
	}
	for i := 3; i < len(t2); i++ {
		if !isNumeric(t2[i]) {
			v2HasPre = true
			break
		}
	}

	// 一个是稳定版本，一个是先行版本
	// 在这个系统中，有后缀的版本大于无后缀的版本
	if !v1HasPre && v2HasPre {
		return -1
	} else if v1HasPre && !v2HasPre {
		return 1
	}

	// 都是先行版本，比较后缀优先级
	if v1HasPre && v2HasPre {
		// 提取后缀部分进行比较
		var v1Suffix, v2Suffix string
		if len(t1) > 3 {
			v1Suffix = strings.Join(t1[3:], "-")
		}
		if len(t2) > 3 {
			v2Suffix = strings.Join(t2[3:], "-")
		}

		priority1 := getSuffixPriority(v1Suffix)
		priority2 := getSuffixPriority(v2Suffix)

		if priority1 != priority2 {
			if priority1 < priority2 {
				return -1
			}
			return 1
		}
	}

	// 继续比较剩余token
	maxLen := len(t1)
	if len(t2) > maxLen {
		maxLen = len(t2)
	}
	for i := 0; i < maxLen; i++ {
		if i >= len(t1) {
			// v1 没有更多token，v2 有
			return -1
		}
		if i >= len(t2) {
			// v2 没有更多token，v1 有
			return 1
		}

		a := t1[i]
		b := t2[i]
		if a == b {
			continue
		}
		aNum, aErr := strconv.Atoi(a)
		bNum, bErr := strconv.Atoi(b)
		if aErr == nil && bErr == nil {
			if aNum < bNum {
				return -1
			}
			return 1
		}
		if aErr == nil && bErr != nil {
			// 数字 > 字母
			return 1
		}
		if aErr != nil && bErr == nil {
			return -1
		}
		cmp := strings.Compare(a, b)
		if cmp != 0 {
			return cmp
		}
	}

	// token完全相同
	return 0
}

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// getSuffixPriority 返回后缀的优先级，数值越大优先级越高
func getSuffixPriority(suffix string) int {
	suffix = strings.ToLower(suffix)
	switch {
	case strings.Contains(suffix, "rc"):
		return 3
	case strings.Contains(suffix, "beta"):
		return 2
	case strings.Contains(suffix, "alpha"):
		return 1
	default:
		return 0
	}
}
