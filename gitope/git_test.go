package gitope

import (
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestParseCommitMessageType(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		// 基本类型测试
		{
			name:     "feat prefix",
			message:  "feat: add new feature",
			expected: "feat",
		},
		{
			name:     "fix prefix",
			message:  "fix: bug fix",
			expected: "fix",
		},
		{
			name:     "refactor prefix",
			message:  "refactor: code refactor",
			expected: "refactor",
		},
		{
			name:     "style prefix",
			message:  "style: format code",
			expected: "style",
		},
		{
			name:     "impr prefix",
			message:  "impr: improve performance",
			expected: "impr",
		},
		{
			name:     "perf prefix",
			message:  "perf: performance improvement",
			expected: "perf",
		},
		{
			name:     "chore prefix",
			message:  "chore: update dependencies",
			expected: "chore",
		},
		{
			name:     "dep prefix",
			message:  "dep: add new dependency",
			expected: "dep",
		},
		{
			name:     "docs prefix",
			message:  "docs: update documentation",
			expected: "docs",
		},
		{
			name:     "test prefix",
			message:  "test: add unit tests",
			expected: "test",
		},
		{
			name:     "typo prefix",
			message:  "typo: fix typo",
			expected: "typo",
		},
		{
			name:     "revert prefix",
			message:  "revert: revert changes",
			expected: "revert",
		},
		{
			name:     "merge prefix",
			message:  "merge: merge branch",
			expected: "merge",
		},
		{
			name:     "wip prefix",
			message:  "wip: work in progress",
			expected: "wip",
		},
		// 大写前缀测试
		{
			name:     "FEAT uppercase prefix",
			message:  "FEAT: add new feature",
			expected: "feat",
		},
		{
			name:     "FIX uppercase prefix",
			message:  "FIX: bug fix",
			expected: "fix",
		},
		// 首字母大写前缀测试
		{
			name:     "Feat title case prefix",
			message:  "Feat: add new feature",
			expected: "feat",
		},
		{
			name:     "Fix title case prefix",
			message:  "Fix: bug fix",
			expected: "fix",
		},
		// 边界情况测试
		{
			name:     "Empty message",
			message:  "",
			expected: "other",
		},
		{
			name:     "Message without prefix",
			message:  "just a regular commit message",
			expected: "other",
		},
		{
			name:     "Message with space after prefix",
			message:  "feat : add new feature",
			expected: "other", // 冒号前有空格，不匹配
		},
		{
			name:     "Message with partial prefix",
			message:  "featu: add new feature",
			expected: "other", // 部分匹配，不匹配
		},
		{
			name:     "Message with longer prefix",
			message:  "featuring: add new feature",
			expected: "other", // 更长前缀，不匹配
		},
		// 特殊字符测试
		{
			name:     "Message with special characters",
			message:  "feat: add new feature!",
			expected: "feat",
		},
		{
			name:     "Message with numbers",
			message:  "feat: add feature v2.0",
			expected: "feat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commit := &object.Commit{
				Message: tt.message,
			}
			result := ParseCommitMessageType(commit)
			if result != tt.expected {
				t.Errorf("ParseCommitMessageType(%q) = %s, want %s", tt.message, result, tt.expected)
			}
		})
	}
}

func TestTagName(t *testing.T) {
	tests := []struct {
		name     string
		refName  string
		expected string
	}{
		{
			name:     "Standard tag ref",
			refName:  "refs/tags/v1.0.0",
			expected: "v1.0.0",
		},
		{
			name:     "Tag without v prefix",
			refName:  "refs/tags/1.0.0",
			expected: "1.0.0",
		},
		{
			name:     "Tag with underscore",
			refName:  "refs/tags/release_1.0.0",
			expected: "release_1.0.0",
		},
		{
			name:     "Tag with dash",
			refName:  "refs/tags/release-1.0.0",
			expected: "release-1.0.0",
		},
		{
			name:     "Tag with dots",
			refName:  "refs/tags/hive.1.0.0",
			expected: "hive.1.0.0",
		},
		{
			name:     "Tag with special characters",
			refName:  "refs/tags/v1.0.0-beta.1",
			expected: "v1.0.0-beta.1",
		},
		{
			name:     "Empty tag name",
			refName:  "refs/tags/",
			expected: "",
		},
		{
			name:     "Tag with numbers only",
			refName:  "refs/tags/123",
			expected: "123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := plumbing.NewReferenceFromStrings(tt.refName, "0000000000000000000000000000000000000000")
			result := TagName(ref)
			if result != tt.expected {
				t.Errorf("TagName(%q) = %s, want %s", tt.refName, result, tt.expected)
			}
		})
	}
}

func TestVersionCompare(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		// 基本版本比较
		{
			name:     "Same versions",
			v1:       "1.0.0",
			v2:       "1.0.0",
			expected: 0,
		},
		{
			name:     "Major version greater",
			v1:       "2.0.0",
			v2:       "1.0.0",
			expected: 1,
		},
		{
			name:     "Major version less",
			v1:       "1.0.0",
			v2:       "2.0.0",
			expected: -1,
		},
		{
			name:     "Minor version greater",
			v1:       "1.2.0",
			v2:       "1.1.0",
			expected: 1,
		},
		{
			name:     "Minor version less",
			v1:       "1.0.2",
			v2:       "1.1.0",
			expected: -1,
		},
		{
			name:     "Minor version less",
			v1:       "1.1.0",
			v2:       "1.2.0",
			expected: -1,
		},
		{
			name:     "Minor version less",
			v1:       "1.0.2-alpha",
			v2:       "1.1.0",
			expected: -1,
		},
		{
			name:     "Minor version less",
			v1:       "1.0.2",
			v2:       "1.1.0-alpha",
			expected: -1,
		},
		{
			name:     "Patch version greater",
			v1:       "1.0.2",
			v2:       "1.0.1",
			expected: 1,
		},
		{
			name:     "Patch version less",
			v1:       "1.0.1",
			v2:       "1.0.2",
			expected: -1,
		},
		// 带v前缀的版本
		{
			name:     "Version with v prefix",
			v1:       "v1.0.0",
			v2:       "1.0.0",
			expected: 0,
		},
		{
			name:     "Both versions with v prefix",
			v1:       "v2.0.0",
			v2:       "v1.0.0",
			expected: 1,
		},
		// 不同长度的版本
		{
			name:     "First version longer",
			v1:       "1.0.0.1",
			v2:       "1.0.0",
			expected: 1,
		},
		{
			name:     "Second version longer",
			v1:       "1.0.0",
			v2:       "1.0.0.1",
			expected: -1,
		},
		// 带后缀的版本
		{
			name:     "Version with alpha suffix",
			v1:       "1.0.0-alpha",
			v2:       "1.0.0",
			expected: 1, // 有后缀大于无后缀
		},
		{
			name:     "Version with alpha suffix",
			v1:       "1.0.0-hotfix",
			v2:       "1.0.0",
			expected: 1, // 有后缀大于无后缀
		},
		{
			name:     "Version with alpha suffix",
			v1:       "1.0.0-hotfix1",
			v2:       "1.0.0",
			expected: 1, // 有后缀大于无后缀
		},
		{
			name:     "Version with alpha suffix",
			v1:       "1.0.0-hotfix1",
			v2:       "1.0.0-hotfix",
			expected: 1, // 有后缀大于无后缀
		},
		{
			name:     "Version with alpha suffix",
			v1:       "1.0.0-hotfix2",
			v2:       "1.0.0-hotfix1",
			expected: 1, // hotfix2 > hotfix1
		},
		{
			name:     "Version with beta suffix",
			v1:       "1.0.0-beta",
			v2:       "1.0.0-alpha",
			expected: 1, // beta > alpha
		},
		{
			name:     "Version with rc suffix",
			v1:       "1.0.0-rc.1",
			v2:       "1.0.0-beta.2",
			expected: 1, // rc > beta
		},
		{
			name:     "Same patch with different suffixes",
			v1:       "1.0.0-rc.2",
			v2:       "1.0.0-rc.1",
			expected: 1, // rc.2 > rc.1
		},
		{
			name:     "Same patch with different suffixes",
			v1:       "1.0.0-alpha1",
			v2:       "1.0.0-alpha2",
			expected: -1, // alpha2 > alpha1
		},
		// 产品名前缀
		{
			name:     "Product prefix hive",
			v1:       "hive.1.0.0",
			v2:       "1.0.0",
			expected: 0, // 前缀被去除，版本相同
		},
		{
			name:     "Product prefix with different versions",
			v1:       "hive.2.0.0",
			v2:       "hive.1.0.0",
			expected: 1,
		},
		// 边界情况
		{
			name:     "Empty versions",
			v1:       "",
			v2:       "",
			expected: 0,
		},
		{
			name:     "First version empty",
			v1:       "",
			v2:       "1.0.0",
			expected: -1,
		},
		{
			name:     "Second version empty",
			v1:       "1.0.0",
			v2:       "",
			expected: 1,
		},
		// 特殊格式
		{
			name:     "Version with underscores",
			v1:       "1_0_0",
			v2:       "1.0.0",
			expected: 0, // 下划线被替换为点
		},
		{
			name:     "Version with plus signs",
			v1:       "1+0+0",
			v2:       "1.0.0",
			expected: 0, // 加号被替换为点
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VersionCompare(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("VersionCompare(%q, %q) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestVersionCompareSymmetry(t *testing.T) {
	// 测试对称性：VersionCompare(a, b) = -VersionCompare(b, a)
	versions := []string{
		"1.0.0",
		"v1.0.0",
		"hive.1.0.0",
		"1.0.1",
		"1.0.2",
		"1.0.2-alpha",
		"1.0.2-beta",
		"1.0.2-beta1",
		"1.0.2-beta2",
		"1.0.2-rc",
		"1.1.0",
		"2.0.0",
		"v2.0.0",
		"hive.2.0.0",
		"2.0.1",
		"2.1.0",
	}

	for i, v1 := range versions {
		for j, v2 := range versions {
			result1 := VersionCompare(v1, v2)
			result2 := VersionCompare(v2, v1)

			// 对称性检查
			if result1 != -result2 {
				t.Errorf("Symmetry check failed: VersionCompare(%q, %q) = %d, VersionCompare(%q, %q) = %d",
					v1, v2, result1, v2, v1, result2)
			}

			// 传递性检查
			if i < j {
				if result1 > 0 {
					t.Errorf("Order check failed: %q should be less than %q, but got %d", v1, v2, result1)
				}
			} else if i > j {
				if result1 < 0 {
					t.Errorf("Order check failed: %q should be greater than %q, but got %d", v1, v2, result1)
				}
			} else {
				if result1 != 0 {
					t.Errorf("Equality check failed: %q should equal %q, but got %d", v1, v2, result1)
				}
			}
		}
	}
}

func TestGetCommitUrl(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		project  string
		hash     string
		expected string
	}{
		{
			name:     "GitLab commit URL",
			base:     "https://gitlab.com",
			project:  "group/project",
			hash:     "abc12345",
			expected: "https://gitlab.com/group/project/-/commits/abc12345",
		},
		{
			name:     "GitHub commit URL",
			base:     "https://github.com",
			project:  "user/repo",
			hash:     "def67890",
			expected: "https://github.com/user/repo/commit/def67890",
		},
		{
			name:     "Generic commit URL",
			base:     "https://example.com",
			project:  "project",
			hash:     "ghi11111",
			expected: "https://example.com/project/commits/ghi11111",
		},
		{
			name:     "HTTP base URL",
			base:     "http://example.com",
			project:  "project",
			hash:     "jkl22222",
			expected: "http://example.com/project/commits/jkl22222",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCommitUrl(tt.base, tt.project, tt.hash)
			if result != tt.expected {
				t.Errorf("GetCommitUrl(%q, %q, %q) = %s, want %s", tt.base, tt.project, tt.hash, result, tt.expected)
			}
		})
	}
}

func TestGetTagUrl(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		project  string
		tagName  string
		expected string
	}{
		{
			name:     "GitLab tag URL",
			base:     "https://gitlab.com",
			project:  "group/project",
			tagName:  "v1.0.0",
			expected: "https://gitlab.com/group/project/-/tags/v1.0.0",
		},
		{
			name:     "GitHub tag URL",
			base:     "https://github.com",
			project:  "user/repo",
			tagName:  "v2.0.0",
			expected: "https://github.com/user/repo/releases/tag/v2.0.0",
		},
		{
			name:     "Generic tag URL",
			base:     "https://example.com",
			project:  "project",
			tagName:  "v3.0.0",
			expected: "https://example.com/project/-/tags/v3.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTagUrl(tt.base, tt.project, tt.tagName)
			if result != tt.expected {
				t.Errorf("GetTagUrl(%q, %q, %q) = %s, want %s", tt.base, tt.project, tt.tagName, result, tt.expected)
			}
		})
	}
}

func TestRenderPipelineUrl(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		project  string
		tagName  string
		expected string
	}{
		{
			name:     "GitLab pipeline URL",
			base:     "https://gitlab.com",
			project:  "group/project",
			tagName:  "v1.0.0",
			expected: "[![CI](https://gitlab.com/group/project/badges/v1.0.0/pipeline.svg?ignore_skipped=true)](https://gitlab.com/group/project/pipelines?page=1&scope=tags&ref=v1.0.0)",
		},
		{
			name:     "GitHub pipeline URL",
			base:     "https://github.com",
			project:  "user/repo",
			tagName:  "v2.0.0",
			expected: "[![GH Action](https://github.com/user/repo/main.yml/badge.svg?branch=v2.0.0)](https://github.com/user/repo/actions)",
		},
		{
			name:     "Unknown platform",
			base:     "https://example.com",
			project:  "project",
			tagName:  "v3.0.0",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderPipelineUrl(tt.base, tt.project, tt.tagName)
			if result != tt.expected {
				t.Errorf("RenderPipelineUrl(%q, %q, %q) = %s, want %s", tt.base, tt.project, tt.tagName, result, tt.expected)
			}
		})
	}
}

// 测试性能
func BenchmarkVersionCompare(b *testing.B) {
	v1 := "hive.2.1.3-beta.2+build.123"
	v2 := "hive.2.1.3-beta.1+build.456"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VersionCompare(v1, v2)
	}
}

func BenchmarkParseCommitMessageType(b *testing.B) {
	commit := &object.Commit{
		Message: "feat: add new feature with long description",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseCommitMessageType(commit)
	}
}

func BenchmarkTagName(b *testing.B) {
	ref := plumbing.NewReferenceFromStrings("refs/tags/v1.0.0-beta.1+build.123", "0000000000000000000000000000000000000000")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TagName(ref)
	}
}

// 测试ChangeLog结构体的方法
func TestChangeLogRenderVersionHeader(t *testing.T) {
	changelog := &ChangeLog{
		Version: "v1.0.0",
		Head: &Ref{
			Hash: "abc12345",
			When: time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC),
		},
	}

	result := changelog.RenderVersionHeader()

	// 验证结果包含必要的信息
	if !strings.Contains(result, "## v1.0.0") {
		t.Errorf("Version header should contain version: %s", result)
	}
	if !strings.Contains(result, "2023-12-25") {
		t.Errorf("Version header should contain date: %s", result)
	}
	if !strings.Contains(result, "abc12345") {
		t.Errorf("Version header should contain hash: %s", result)
	}
}

func TestChangeLogRenderCommit(t *testing.T) {
	changelog := &ChangeLog{
		Version: "v1.0.0",
		Head: &Ref{
			Hash: "abc12345",
			When: time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC),
		},
	}

	commit := &object.Commit{
		Hash:    plumbing.NewHash("abc12345"),
		Message: "feat: add new feature\n\nThis is a detailed description\nof the new feature.",
		Author: object.Signature{
			Name:  "John Doe",
			Email: "john@example.com",
			When:  time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC),
		},
	}

	result := changelog.RenderCommit(commit)

	// 验证结果包含必要的信息
	if !strings.Contains(result, "feat: add new feature") {
		t.Errorf("Commit message should contain commit message: %s", result)
	}
	if !strings.Contains(result, "abc12345") {
		t.Errorf("Commit message should contain hash: %s", result)
	}
	if !strings.Contains(result, "John Doe") {
		t.Errorf("Commit message should contain author name: %s", result)
	}
	if !strings.Contains(result, "This is a detailed description") {
		t.Errorf("Commit message should contain detailed description: %s", result)
	}
}
