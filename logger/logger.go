package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	warnLogger  *log.Logger
	file        *os.File
}

// NewLogger 创建新的日志记录器
func NewLogger(logFile string) (*Logger, error) {
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("无法打开日志文件: %v", err)
	}

	return &Logger{
		infoLogger:  log.New(file, "[INFO] ", log.LstdFlags),
		errorLogger: log.New(file, "[ERROR] ", log.LstdFlags),
		warnLogger:  log.New(file, "[WARN] ", log.LstdFlags),
		file:        file,
	}, nil
}

// Info 记录信息日志
func (l *Logger) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.infoLogger.Println(msg)
	fmt.Printf("[INFO] %s %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
}

// Error 记录错误日志
func (l *Logger) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.errorLogger.Println(msg)
	fmt.Printf("[ERROR] %s %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
}

// Warn 记录警告日志
func (l *Logger) Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.warnLogger.Println(msg)
	fmt.Printf("[WARN] %s %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
}

// Close 关闭日志文件
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
} 