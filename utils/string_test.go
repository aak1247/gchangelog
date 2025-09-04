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

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		version1 string
		version2 string
		expected int
	}{
		// 基本版本比较
		{
			name:     "Same versions",
			version1: "1.0.0",
			version2: "1.0.0",
			expected: 0,
		},
		{
			name:     "First version greater - major",
			version1: "2.0.0",
			version2: "1.0.0",
			expected: 1,
		},
		{
			name:     "First version greater - minor",
			version1: "1.2.0",
			version2: "1.1.0",
			expected: 1,
		},
		{
			name:     "First version greater - patch",
			version1: "1.0.2",
			version2: "1.0.1",
			expected: 1,
		},
		{
			name:     "Second version greater - major",
			version1: "1.0.0",
			version2: "2.0.0",
			expected: -1,
		},
		{
			name:     "Second version greater - minor",
			version1: "1.1.0",
			version2: "1.2.0",
			expected: -1,
		},
		{
			name:     "Second version greater - patch",
			version1: "1.0.1",
			version2: "1.0.2",
			expected: -1,
		},
		// 不同长度的版本
		{
			name:     "First version longer",
			version1: "1.0.0.1",
			version2: "1.0.0",
			expected: 1,
		},
		{
			name:     "Second version longer",
			version1: "1.0.0",
			version2: "1.0.0.1",
			expected: -1,
		},
		// 包含字母的版本
		{
			name:     "Alpha versions",
			version1: "1.0.0-alpha",
			version2: "1.0.0-beta",
			expected: -1, // alpha < beta lexicographically
		},
		{
			name:     "Mixed numeric and alpha",
			version1: "1.0.0-rc1",
			version2: "1.0.0-rc2",
			expected: -1, // rc1 < rc2 lexicographically
		},
		// 边界情况
		{
			name:     "Empty versions",
			version1: "",
			version2: "",
			expected: 0,
		},
		{
			name:     "First version empty",
			version1: "",
			version2: "1.0.0",
			expected: -1,
		},
		{
			name:     "Second version empty",
			version1: "1.0.0",
			version2: "",
			expected: 1,
		},
		// 复杂版本格式
		{
			name:     "Complex version comparison",
			version1: "2.1.3-beta.2",
			version2: "2.1.3-beta.1",
			expected: 1,
		},
		{
			name:     "Complex version comparison",
			version1: "hive.0.0.4-alpha8",
			version2: "hive.0.0.4-alpha9",
			expected: -1,
		},
		{
			name:     "Complex version comparison",
			version1: "hive.0.0.4-alpha",
			version2: "hive.0.0.4-alpha9",
			expected: -1,
		},
		{
			name:     "Complex version comparison",
			version1: "hive.0.0.4-alpha8",
			version2: "hive.0.0.5-alpha9",
			expected: -1,
		},
		{
			name:     "Version with build number",
			version1: "1.0.0+build.123",
			version2: "1.0.0+build.456",
			expected: 0, // build numbers are typically ignored in version comparison
		},
		// 特殊字符
		{
			name:     "Version with dots only",
			version1: "1.2",
			version2: "1.2.0",
			expected: -1,
		},
		{
			name:     "Version with leading zeros",
			version1: "1.02.0",
			version2: "1.2.0",
			expected: 0, // 02 and 2 are treated as same numeric value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareVersions(tt.version1, tt.version2)
			if result != tt.expected {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.version1, tt.version2, result, result)
			}
		})
	}
}

// 测试CompareVersions的对称性
func TestCompareVersionsSymmetry(t *testing.T) {
	versions := []string{
		"1.0.0",
		"1.0.1",
		"1.1.0",
		"2.0.0",
		"2.0.1",
		"2.1.0",
	}

	for i, v1 := range versions {
		for j, v2 := range versions {
			result1 := CompareVersions(v1, v2)
			result2 := CompareVersions(v2, v1)

			// 对称性检查：CompareVersions(a, b) = -CompareVersions(b, a)
			if result1 != -result2 {
				t.Errorf("Symmetry check failed: CompareVersions(%q, %q) = %d, CompareVersions(%q, %q) = %d",
					v1, v2, result1, v2, v1, result2)
			}

			// 传递性检查：如果 a < b 且 b < c，那么 a < c
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

// 测试CompareVersions的性能
func BenchmarkCompareVersions(b *testing.B) {
	version1 := "2.1.3-beta.2+build.123"
	version2 := "2.1.3-beta.1+build.456"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CompareVersions(version1, version2)
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
