package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	logMu   sync.Mutex
	logFile *os.File
)

// initLogger 打开日志文件（追加模式，不存在则创建）
func initLogger(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	logFile = f
	return nil
}

// closeLogger 关闭日志文件
func closeLogger() {
	logMu.Lock()
	defer logMu.Unlock()
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
	}
}

// logAction 写一行操作日志：时间 | 用户名 | 内容，同时输出到控制台与日志文件
func logAction(username, msg string) {
	line := fmt.Sprintf("%s | %s | %s", time.Now().Format("2006-01-02 15:04:05"), username, msg)
	logMu.Lock()
	defer logMu.Unlock()
	if logFile != nil {
		fmt.Fprintln(logFile, line)
	}
	fmt.Println(line)
}
