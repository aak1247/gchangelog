package utils

import (
	"testing"
)

func TestIsMultiline(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Single line string",
			input:    "Hello World",
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "String with newline at end",
			input:    "Hello World\n",
			expected: true,
		},
		{
			name:     "String with newline in middle",
			input:    "Hello\nWorld",
			expected: true,
		},
		{
			name:     "String with multiple newlines",
			input:    "Hello\n\nWorld",
			expected: true,
		},
		{
			name:     "String with carriage return and newline",
			input:    "Hello\r\nWorld",
			expected: true,
		},
		{
			name:     "String with only newline",
			input:    "\n",
			expected: true,
		},
		{
			name:     "String with unicode newline",
			input:    "Hello\u000AWorld",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMultiline(tt.input)
			if result != tt.expected {
				t.Errorf("IsMultiline(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// 测试IsMultiline的性能
func BenchmarkIsMultiline(b *testing.B) {
	testString := "This is a test string with\nmultiple lines\nfor benchmarking"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsMultiline(testString)
	}
}
