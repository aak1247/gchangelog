package main

import (
	"flag"
	"github.com/aak1247/gchangelog/configs"
	"github.com/aak1247/gchangelog/gitope"
	"github.com/aak1247/gchangelog/utils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"log"
	"time"
)

func main() {
	var repoPath string
	var fileName = "changelog.md"
	var skip = "skip"
	flag.StringVar(&repoPath, "p", "", "仓库地址")
	flag.StringVar(&fileName, "f", "changelog.md", "文件名")
	flag.StringVar(&skip, "skip", "skip", "跳过消息")
	flag.BoolVar(&configs.MR, "mr", false, "记录mr日志")
	flag.Parse()
	if repoPath == "" {
		panic("repo path not presented")
	}
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		panic(err)
	}
	if err = configs.ParseSkipMsg(skip); err != nil {
		panic(err)
	}

	configs.Project = gitope.GetProjectPath(r) // 用于生成链接
	configs.BaseUrl = gitope.GetBaseUrl(r)

	tag1, tag2, err := gitope.FindTag(err, r)
	if err != nil {
		log.Fatal("failed to find tag", err)
	}
	commits := gitope.FindCommits(tag2, tag1, r)
	head := &gitope.Ref{}
	commit, err := r.CommitObject(tag1.Hash())
	if err != nil {
		log.Println("failed to find commits", err)
		head.Hash = tag1.Hash().String()
		tag, _ := r.TagObject(tag1.Hash())
		if tag != nil {
			head.When = tag.Tagger.When
		} else {
			head.When = time.Now()
		}
	} else {
		head.Hash = commit.Hash.String()
		head.When = commit.Author.When
	}
	// 解析commit msg，然后生成新增changelog
	version := gitope.TagName(tag1)
	res := &gitope.ChangeLog{
		Version: version,
		Head: &gitope.Ref{
			Hash: head.Hash,
			When: head.When,
		},
		Groups: make(map[string][]*object.Commit),
	}
	res.ParseCommits(commits)
	// 拼接输出
	str := res.String()

	// 输出
	log.Println("Generated changelog:\n", str)
	// 写入文件
	err = utils.InsertToFile(fileName, str, configs.ChangelogHeaderLines)
	if err != nil {
		log.Fatal(err)
	}
}
