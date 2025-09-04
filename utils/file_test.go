package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileExists(t *testing.T) {
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "file_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test_file.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Existing file",
			path:     testFile,
			expected: true,
		},
		{
			name:     "Non-existing file",
			path:     filepath.Join(tempDir, "non_existing.txt"),
			expected: false,
		},
		{
			name:     "Empty path",
			path:     "",
			expected: false,
		},
		{
			name:     "Directory path",
			path:     tempDir,
			expected: true, // 目录也存在
		},
		{
			name:     "Relative path",
			path:     "./non_existing_file.txt",
			expected: false,
		},
		{
			name:     "Absolute path to non-existing",
			path:     "/tmp/non_existing_file_12345.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FileExists(tt.path)
			if result != tt.expected {
				t.Errorf("FileExists(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestInsertToFile(t *testing.T) {
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "insert_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		content     string
		skipRows    int
		expectError bool
		setupFile   func(string) // 设置测试文件的函数
		validate    func(string) // 验证结果的函数
	}{
		{
			name:     "Insert at beginning of new file",
			content:  "New content\nSecond line",
			skipRows: 0,
			setupFile: func(path string) {
				// 不创建文件，让函数自动创建
			},
			validate: func(path string) {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Errorf("Failed to read file: %v", err)
					return
				}
				content := string(data)
				if !strings.Contains(content, "New content") {
					t.Errorf("Expected content not found in file")
				}
				if !strings.Contains(content, "Second line") {
					t.Errorf("Expected second line not found in file")
				}
			},
		},
		{
			name:     "Insert at beginning of existing file",
			content:  "Inserted content",
			skipRows: 0,
			setupFile: func(path string) {
				err := os.WriteFile(path, []byte("Original content\nSecond line"), 0o644)
				if err != nil {
					t.Errorf("Failed to setup test file: %v", err)
				}
			},
			validate: func(path string) {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Errorf("Failed to read file: %v", err)
					return
				}
				content := string(data)
				lines := strings.Split(content, "\n")
				if len(lines) < 3 {
					t.Errorf("Expected at least 3 lines, got %d", len(lines))
					return
				}
				if lines[0] != "Inserted content" {
					t.Errorf("First line should be 'Inserted content', got '%s'", lines[0])
				}
				if lines[1] != "Original content" {
					t.Errorf("Second line should be 'Original content', got '%s'", lines[1])
				}
			},
		},
		{
			name:     "Insert in middle of file",
			content:  "Middle content",
			skipRows: 1,
			setupFile: func(path string) {
				err := os.WriteFile(path, []byte("First line\nSecond line\nThird line"), 0o644)
				if err != nil {
					t.Errorf("Failed to setup test file: %v", err)
				}
			},
			validate: func(path string) {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Errorf("Failed to read file: %v", err)
					return
				}
				content := string(data)
				lines := strings.Split(content, "\n")
				if len(lines) < 4 {
					t.Errorf("Expected at least 4 lines, got %d", len(lines))
					return
				}
				if lines[0] != "First line" {
					t.Errorf("First line should be 'First line', got '%s'", lines[0])
				}
				if lines[1] != "Middle content" {
					t.Errorf("Second line should be 'Middle content', got '%s'", lines[1])
				}
				if lines[2] != "Second line" {
					t.Errorf("Third line should be 'Second line', got '%s'", lines[2])
				}
			},
		},
		{
			name:     "Insert at end of file",
			content:  "End content",
			skipRows: 2,
			setupFile: func(path string) {
				err := os.WriteFile(path, []byte("First line\nSecond line"), 0o644)
				if err != nil {
					t.Errorf("Failed to setup test file: %v", err)
				}
			},
			validate: func(path string) {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Errorf("Failed to read file: %v", err)
					return
				}
				content := string(data)
				lines := strings.Split(content, "\n")
				if len(lines) < 3 {
					t.Errorf("Expected at least 3 lines, got %d", len(lines))
					return
				}
				if lines[2] != "End content" {
					t.Errorf("Third line should be 'End content', got '%s'", lines[2])
				}
			},
		},
		{
			name:     "Insert with skipRows exceeding file length",
			content:  "Content",
			skipRows: 10,
			setupFile: func(path string) {
				err := os.WriteFile(path, []byte("First line\nSecond line"), 0o644)
				if err != nil {
					t.Errorf("Failed to setup test file: %v", err)
				}
			},
			validate: func(path string) {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Errorf("Failed to read file: %v", err)
					return
				}
				content := string(data)
				lines := strings.Split(content, "\n")
				if len(lines) < 3 {
					t.Errorf("Expected at least 3 lines, got %d", len(lines))
					return
				}
				if lines[2] != "Content" {
					t.Errorf("Last line should be 'Content', got '%s'", lines[2])
				}
			},
		},
		{
			name:     "Insert empty content",
			content:  "",
			skipRows: 0,
			setupFile: func(path string) {
				err := os.WriteFile(path, []byte("Original content"), 0o644)
				if err != nil {
					t.Errorf("Failed to setup test file: %v", err)
				}
			},
			validate: func(path string) {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Errorf("Failed to read file: %v", err)
					return
				}
				content := string(data)
				lines := strings.Split(content, "\n")
				if len(lines) < 2 {
					t.Errorf("Expected at least 2 lines, got %d", len(lines))
					return
				}
				if lines[0] != "" {
					t.Errorf("First line should be empty, got '%s'", lines[0])
				}
				if lines[1] != "Original content" {
					t.Errorf("Second line should be 'Original content', got '%s'", lines[1])
				}
			},
		},
		{
			name:     "Insert content with multiple newlines",
			content:  "Line1\n\nLine3",
			skipRows: 0,
			setupFile: func(path string) {
				// 不创建文件，让函数自动创建
			},
			validate: func(path string) {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Errorf("Failed to read file: %v", err)
					return
				}
				content := string(data)
				lines := strings.Split(content, "\n")
				if len(lines) < 4 {
					t.Errorf("Expected at least 4 lines, got %d", len(lines))
					return
				}
				if lines[0] != "Line1" {
					t.Errorf("First line should be 'Line1', got '%s'", lines[0])
				}
				if lines[1] != "" {
					t.Errorf("Second line should be empty, got '%s'", lines[1])
				}
				if lines[2] != "Line3" {
					t.Errorf("Third line should be 'Line3', got '%s'", lines[2])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tempDir, "test_insert_"+tt.name+".txt")

			// 设置测试文件
			if tt.setupFile != nil {
				tt.setupFile(testFile)
			}

			// 执行插入操作
			err := InsertToFile(testFile, tt.content, tt.skipRows)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// 验证结果
			if tt.validate != nil {
				tt.validate(testFile)
			}
		})
	}
}

// 测试文件权限
func TestInsertToFilePermissions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "perm_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "perm_test.txt")

	// 测试创建新文件时的权限
	err = InsertToFile(testFile, "test content", 0)
	if err != nil {
		t.Fatalf("Failed to insert to file: %v", err)
	}

	// 检查文件权限
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	// 检查文件是否可读
	if info.Mode()&0o400 == 0 {
		t.Errorf("File should be readable")
	}
}

// 测试并发插入
func TestInsertToFileConcurrent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "concurrent_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "concurrent_test.txt")

	// 创建多个goroutine同时插入内容
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() { done <- true }()
			content := fmt.Sprintf("Content from goroutine %d", id)
			err := InsertToFile(testFile, content, id)
			if err != nil {
				t.Errorf("Goroutine %d failed: %v", id, err)
			}
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 5; i++ {
		<-done
	}

	// 验证文件内容
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	content := string(data)
	for i := 0; i < 5; i++ {
		expected := fmt.Sprintf("Content from goroutine %d", i)
		if !strings.Contains(content, expected) {
			t.Errorf("Expected content '%s' not found", expected)
		}
	}
}

// 性能测试
func BenchmarkFileExists(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "benchmark")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "benchmark.txt")
	err = os.WriteFile(testFile, []byte("test"), 0o644)
	if err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FileExists(testFile)
	}
}

func BenchmarkInsertToFile(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "benchmark")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "benchmark.txt")
	content := strings.Repeat("This is a test line\n", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := InsertToFile(testFile, content, 0)
		if err != nil {
			b.Fatalf("Failed to insert: %v", err)
		}
	}
}
