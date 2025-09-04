package utils

import (
	"bufio"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/aak1247/gchangelog/configs"
)

// fileMutexes provides per-file mutexes to make InsertToFile concurrency-safe
var fileMutexes struct {
	sync.Mutex
	m map[string]*sync.Mutex
}

func init() {
	fileMutexes.m = make(map[string]*sync.Mutex)
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func InsertToFile(path string, content string, skipRows int) error {
	// Ensure concurrency-safety per file using a process-wide mutex map
	fileMutexes.Lock()
	mu, ok := fileMutexes.m[path]
	if !ok {
		mu = &sync.Mutex{}
		fileMutexes.m[path] = mu
	}
	fileMutexes.Unlock()

	mu.Lock()
	defer mu.Unlock()
	if !FileExists(path) {
		// 创建文件
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		// 写入文件头
		file.WriteString(configs.DefaultHead)
		file.Close()
	}
	// 打开文件以读取
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// 读取文件的所有内容
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// 确保 skipRows 不超过文件的行数
	if skipRows > len(lines) {
		skipRows = len(lines)
	}

	// 将内容插入到指定位置
	newLines := append(lines[:skipRows], append(strings.Split(content, "\n"), lines[skipRows:]...)...)

	// 将新的内容写回到文件中
	output := strings.Join(newLines, "\n")
	err = ioutil.WriteFile(path, []byte(output), 0644)
	if err != nil {
		return err
	}

	return nil
}
