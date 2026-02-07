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
	// 收集所有符合条件的tags
	var allTags []*plumbing.Reference
	tagIter, err := r.Tags()
	if err != nil {
		panic(err)
	}

	// 收集所有tags
	for {
		tag, err := tagIter.Next()
		if err != nil || tag == nil {
			break
		}

		// 检查tag是否在时间范围内
		if isTagRecent(r, tag) {
			allTags = append(allTags, tag)
		}
	}

	// 如果没有tags
	if len(allTags) == 0 {
		panic("no tag found")
	}

	// 如果有数量限制，只保留最近的N个tags
	if configs.MaxTagCount > 0 && len(allTags) > configs.MaxTagCount {
		// 按时间排序，取最新的N个
		sortedTags := make([]*plumbing.Reference, len(allTags))
		copy(sortedTags, allTags)

		// 按时间从新到旧排序
		for i := 0; i < len(sortedTags)-1; i++ {
			for j := i + 1; j < len(sortedTags); j++ {
				time1, _ := getTagTime(r, sortedTags[i])
				time2, _ := getTagTime(r, sortedTags[j])
				if time1.Before(time2) {
					sortedTags[i], sortedTags[j] = sortedTags[j], sortedTags[i]
				}
			}
		}

		allTags = sortedTags[:configs.MaxTagCount]
	}

	// 在符合条件的tags中找到版本最大的两个
	var tag1, tag2 *plumbing.Reference
	var tag1Name string

	if len(allTags) > 0 {
		tag1 = allTags[0]
		tag1Name = TagName(tag1)
	}

	for i := 1; i < len(allTags); i++ {
		tagN := allTags[i]
		tagNName := TagName(tagN)
		if VersionCompare(tagNName, tag1Name) > 0 {
			tag1Name = TagName(tag1)
			tag2 = tag1
			tag1 = tagN
		} else if tag2 == nil || VersionCompare(tagNName, TagName(tag2)) > 0 {
			tag2 = tagN
		}
	}

	return tag1, tag2, err
}

func FindPreviousTag(r *git.Repository, currentTag *plumbing.Reference) (*plumbing.Reference, error) {
	// 收集所有符合条件的tags
	var allTags []*plumbing.Reference
	var currentTagName = TagName(currentTag)
	tagIter, err := r.Tags()
	if err != nil {
		panic(err)
	}

	// 收集所有tags
	for {
		tag, err := tagIter.Next()
		if err != nil || tag == nil {
			break
		}

		// 检查tag是否在时间范围内
		if isTagRecent(r, tag) {
			allTags = append(allTags, tag)
		}
	}

	// 如果有数量限制，只保留最近的N个tags
	if configs.MaxTagCount > 0 && len(allTags) > configs.MaxTagCount {
		// 按时间排序，取最新的N个
		sortedTags := make([]*plumbing.Reference, len(allTags))
		copy(sortedTags, allTags)

		// 按时间从新到旧排序
		for i := 0; i < len(sortedTags)-1; i++ {
			for j := i + 1; j < len(sortedTags); j++ {
				time1, _ := getTagTime(r, sortedTags[i])
				time2, _ := getTagTime(r, sortedTags[j])
				if time1.Before(time2) {
					sortedTags[i], sortedTags[j] = sortedTags[j], sortedTags[i]
				}
			}
		}

		allTags = sortedTags[:configs.MaxTagCount]
	}

	// 在符合条件的tags中找到小于当前tag的最大版本
	var resultTag *plumbing.Reference
	var resultTagName = "0.0.0"

	for _, tag := range allTags {
		tagName := TagName(tag)
		if VersionCompare(tagName, resultTagName) > 0 && VersionCompare(tagName, currentTagName) < 0 {
			resultTag = tag
			resultTagName = tagName
		}
	}

	if resultTag == nil {
		// 没有找到符合条件的tag
		return nil, err
	}

	return resultTag, err
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

type suffixKind int

const (
	suffixPrerelease suffixKind = iota
	suffixNone
	suffixPostrelease
)

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

	rankByOrder := func(s string, order []string) int {
		s = strings.ToLower(s)
		for i, token := range order {
			if token == "" {
				continue
			}
			if strings.Contains(s, token) {
				return i + 1
			}
		}
		return 0
	}

	suffixKindOf := func(suffix string) suffixKind {
		suffix = strings.ToLower(suffix)
		if strings.TrimSpace(suffix) == "" {
			return suffixNone
		}
		if rankByOrder(suffix, configs.PostreleaseSuffixOrder) > 0 {
			return suffixPostrelease
		}
		if rankByOrder(suffix, configs.PrereleaseSuffixOrder) > 0 {
			return suffixPrerelease
		}
		// 未识别的后缀默认按“正式版后”处理（大于无后缀正式版）
		return suffixPostrelease
	}

	// 以点或短横分割所有token（短横用于先行版本）
	sep := regexp.MustCompile(`[\.-]`)
	t1 := sep.Split(v1, -1)
	t2 := sep.Split(v2, -1)

	numericPrefixLen := func(tokens []string) int {
		for i := 0; i < len(tokens); i++ {
			if _, err := strconv.Atoi(tokens[i]); err != nil {
				return i
			}
		}
		return len(tokens)
	}

	n1 := numericPrefixLen(t1)
	n2 := numericPrefixLen(t2)

	// 比较纯数字前缀（支持 1.0.0.1 这类多段数字版本）
	for i := 0; i < n1 && i < n2; i++ {
		aNum, _ := strconv.Atoi(t1[i])
		bNum, _ := strconv.Atoi(t2[i])
		if aNum < bNum {
			return -1
		}
		if aNum > bNum {
			return 1
		}
	}

	// 数字前缀完全相同，数字段更长者更大：1.0.0.1 > 1.0.0
	if n1 < n2 {
		return -1
	}
	if n1 > n2 {
		return 1
	}

	suffix1 := strings.Join(t1[n1:], "-")
	suffix2 := strings.Join(t2[n2:], "-")
	kind1 := suffixKindOf(suffix1)
	kind2 := suffixKindOf(suffix2)

	// 数字版本完全相同：后缀按语义排序 prerelease < none < postrelease
	if kind1 != kind2 {
		if kind1 < kind2 {
			return -1
		}
		return 1
	}

	// 都是 prerelease，比较后缀优先级（rc > beta > alpha）
	if kind1 == suffixPrerelease {
		priority1 := getSuffixPriority(suffix1)
		priority2 := getSuffixPriority(suffix2)
		if priority1 > 0 && priority2 > 0 && priority1 != priority2 {
			if priority1 < priority2 {
				return -1
			}
			return 1
		}
	}

	// 都是 postrelease，比较后缀优先级（按配置表排序）
	if kind1 == suffixPostrelease {
		priority1 := getPostSuffixPriority(suffix1)
		priority2 := getPostSuffixPriority(suffix2)
		if priority1 > 0 && priority2 > 0 && priority1 != priority2 {
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

// getTagTime 获取tag的创建时间
func getTagTime(r *git.Repository, tagRef *plumbing.Reference) (time.Time, error) {
	if tagRef == nil {
		return time.Time{}, fmt.Errorf("tag reference is nil")
	}

	// 尝试获取tag对象
	tagObj, err := r.TagObject(tagRef.Hash())
	if err == nil {
		return tagObj.Tagger.When, nil
	}

	// 如果不是annotated tag，尝试获取commit对象
	commit, err := r.CommitObject(tagRef.Hash())
	if err == nil {
		return commit.Author.When, nil
	}

	return time.Time{}, fmt.Errorf("cannot get time for tag %s", tagRef.Name().String())
}

// isTagRecent 检查tag是否在指定的时间范围内
func isTagRecent(r *git.Repository, tagRef *plumbing.Reference) bool {
	if !configs.OnlyRecentTags {
		return true
	}

	tagTime, err := getTagTime(r, tagRef)
	if err != nil {
		// 如果无法获取时间，默认包含
		return true
	}

	// 检查是否超过最大天数
	if configs.MaxTagAgeDays > 0 {
		cutoff := time.Now().AddDate(0, 0, -configs.MaxTagAgeDays)
		if tagTime.Before(cutoff) {
			return false
		}
	}

	return true
}

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// getSuffixPriority 返回后缀的优先级，数值越大优先级越高
func getSuffixPriority(suffix string) int {
	suffix = strings.ToLower(suffix)
	for i, token := range configs.PrereleaseSuffixOrder {
		if token == "" {
			continue
		}
		if strings.Contains(suffix, token) {
			return i + 1
		}
	}
	return 0
}

func getPostSuffixPriority(suffix string) int {
	suffix = strings.ToLower(suffix)
	for i, token := range configs.PostreleaseSuffixOrder {
		if token == "" {
			continue
		}
		if strings.Contains(suffix, token) {
			return i + 1
		}
	}
	return 0
}
