package configs

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseSkipMsg(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedLen int
		expected    []string
	}{
		{
			name:        "Single message",
			input:       "skip this",
			expectedLen: 1,
			expected:    []string{"skip this"},
		},
		{
			name:        "Multiple messages with comma",
			input:       "skip1,skip2,skip3",
			expectedLen: 3,
			expected:    []string{"skip1", "skip2", "skip3"},
		},
		{
			name:        "Empty string",
			input:       "",
			expectedLen: 1,
			expected:    []string{""},
		},
		{
			name:        "Single comma",
			input:       ",",
			expectedLen: 2,
			expected:    []string{"", ""},
		},
		{
			name:        "Multiple commas",
			input:       ",,,",
			expectedLen: 4,
			expected:    []string{"", "", "", ""},
		},
		{
			name:        "Messages with spaces",
			input:       " skip1 , skip2 , skip3 ",
			expectedLen: 3,
			expected:    []string{" skip1 ", " skip2 ", " skip3 "},
		},
		{
			name:        "Complex messages",
			input:       "feat: new feature,fix: bug fix,refactor: code refactor",
			expectedLen: 3,
			expected:    []string{"feat: new feature", "fix: bug fix", "refactor: code refactor"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置全局变量
			SkipMsgs = make([]string, 0)

			err := ParseSkipMsg(tt.input)
			if err != nil {
				t.Errorf("ParseSkipMsg(%q) returned error: %v", tt.input, err)
			}

			if len(SkipMsgs) != tt.expectedLen {
				t.Errorf("ParseSkipMsg(%q) expected length %d, got %d", tt.input, tt.expectedLen, len(SkipMsgs))
			}

			// 验证内容
			for i, expected := range tt.expected {
				if i < len(SkipMsgs) && SkipMsgs[i] != expected {
					t.Errorf("ParseSkipMsg(%q) at index %d expected %q, got %q", tt.input, i, expected, SkipMsgs[i])
				}
			}
		})
	}
}

func TestParseSkipMsgAccumulation(t *testing.T) {
	// 测试多次调用是否会累积消息
	SkipMsgs = make([]string, 0)

	// 第一次调用
	err := ParseSkipMsg("first,second")
	if err != nil {
		t.Fatalf("First ParseSkipMsg failed: %v", err)
	}
	if len(SkipMsgs) != 2 {
		t.Errorf("Expected 2 messages after first call, got %d", len(SkipMsgs))
	}

	// 第二次调用
	err = ParseSkipMsg("third,fourth")
	if err != nil {
		t.Fatalf("Second ParseSkipMsg failed: %v", err)
	}
	if len(SkipMsgs) != 4 {
		t.Errorf("Expected 4 messages after second call, got %d", len(SkipMsgs))
	}

	// 验证累积的内容
	expected := []string{"first", "second", "third", "fourth"}
	for i, exp := range expected {
		if i < len(SkipMsgs) && SkipMsgs[i] != exp {
			t.Errorf("At index %d expected %q, got %q", i, exp, SkipMsgs[i])
		}
	}
}

func TestSkipsShouldSkip(t *testing.T) {
	tests := []struct {
		name       string
		skipMsgs   []string
		message    string
		shouldSkip bool
	}{
		{
			name:       "Exact match",
			skipMsgs:   []string{"skip this", "ignore that"},
			message:    "skip this",
			shouldSkip: true,
		},
		{
			name:       "No match",
			skipMsgs:   []string{"skip this", "ignore that"},
			message:    "keep this",
			shouldSkip: false,
		},
		{
			name:       "Match with leading space",
			skipMsgs:   []string{"skip this", "ignore that"},
			message:    " skip this",
			shouldSkip: true,
		},
		{
			name:       "Match with trailing space",
			skipMsgs:   []string{"skip this", "ignore that"},
			message:    "skip this ",
			shouldSkip: true,
		},
		{
			name:       "Match with both spaces",
			skipMsgs:   []string{"skip this", "ignore that"},
			message:    " skip this ",
			shouldSkip: true,
		},
		{
			name:       "Empty skip message",
			skipMsgs:   []string{"", "skip this"},
			message:    "",
			shouldSkip: true,
		},
		{
			name:       "Empty message with non-empty skips",
			skipMsgs:   []string{"skip this"},
			message:    "",
			shouldSkip: false,
		},
		{
			name:       "Case sensitive match",
			skipMsgs:   []string{"Skip This", "IGNORE THAT"},
			message:    "skip this",
			shouldSkip: false, // 应该区分大小写
		},
		{
			name:       "Partial match should not skip",
			skipMsgs:   []string{"skip"},
			message:    "skip this message",
			shouldSkip: false, // 应该是完全匹配
		},
		{
			name:       "Empty skip list",
			skipMsgs:   []string{},
			message:    "any message",
			shouldSkip: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skips := skips(tt.skipMsgs)
			result := skips.ShouldSkip(tt.message)

			if result != tt.shouldSkip {
				t.Errorf("ShouldSkip(%q) with skips %v expected %v, got %v",
					tt.message, tt.skipMsgs, tt.shouldSkip, result)
			}
		})
	}
}

func TestSkipsShouldSkipWithRealData(t *testing.T) {
	// 使用真实的提交消息格式进行测试
	skipMsgs := []string{
		"chore: update dependencies",
		"docs: update README",
		"style: format code",
		"test: add unit tests",
	}

	skips := skips(skipMsgs)

	realMessages := []struct {
		message    string
		shouldSkip bool
	}{
		{"chore: update dependencies", true},
		{"docs: update README", true},
		{"style: format code", true},
		{"test: add unit tests", true},
		{"feat: new feature", false},
		{"fix: bug fix", false},
		{"chore: update dependencies ", true},          // 尾随空格
		{" chore: update dependencies", true},          // 前导空格
		{"Chore: update dependencies", false},          // 大小写不匹配
		{"chore: update dependencies and more", false}, // 部分匹配
	}

	for _, rm := range realMessages {
		result := skips.ShouldSkip(rm.message)
		if result != rm.shouldSkip {
			t.Errorf("ShouldSkip(%q) expected %v, got %v", rm.message, rm.shouldSkip, result)
		}
	}
}

// 测试性能
func BenchmarkParseSkipMsg(b *testing.B) {
	message := "skip1,skip2,skip3,skip4,skip5,skip6,skip7,skip8,skip9,skip10"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SkipMsgs = make([]string, 0)
		ParseSkipMsg(message)
	}
}

func BenchmarkShouldSkip(b *testing.B) {
	skipMsgs := []string{
		"chore: update dependencies",
		"docs: update README",
		"style: format code",
		"test: add unit tests",
		"feat: new feature",
		"fix: bug fix",
		"refactor: code refactor",
		"perf: performance improvement",
	}

	skips := skips(skipMsgs)
	message := "chore: update dependencies"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		skips.ShouldSkip(message)
	}
}

// 测试边界情况
func TestParseSkipMsgEdgeCases(t *testing.T) {
	// 测试非常长的消息
	longMessage := strings.Repeat("skip,", 1000) + "end"
	SkipMsgs = make([]string, 0)

	err := ParseSkipMsg(longMessage)
	if err != nil {
		t.Errorf("ParseSkipMsg with long message failed: %v", err)
	}

	if len(SkipMsgs) != 1001 {
		t.Errorf("Expected 1001 messages, got %d", len(SkipMsgs))
	}

	// 测试包含特殊字符的消息
	specialMessage := "skip\nthis,skip\tthat,skip\r\nthis"
	SkipMsgs = make([]string, 0)

	err = ParseSkipMsg(specialMessage)
	if err != nil {
		t.Errorf("ParseSkipMsg with special characters failed: %v", err)
	}

	if len(SkipMsgs) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(SkipMsgs))
	}
}

// 测试并发安全性
func TestParseSkipMsgConcurrent(t *testing.T) {
	SkipMsgs = make([]string, 0)

	// 创建多个goroutine同时调用ParseSkipMsg
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() { done <- true }()
			message := fmt.Sprintf("skip%d,ignore%d", id, id)
			err := ParseSkipMsg(message)
			if err != nil {
				t.Errorf("Goroutine %d failed: %v", id, err)
			}
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 5; i++ {
		<-done
	}

	// 验证结果（注意：由于并发，顺序可能不确定）
	if len(SkipMsgs) != 10 {
		t.Errorf("Expected 10 messages, got %d", len(SkipMsgs))
	}
}
