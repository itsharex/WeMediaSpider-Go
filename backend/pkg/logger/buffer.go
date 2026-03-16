package logger

import (
	"sync"
)

// LogBuffer 日志缓冲区
type LogBuffer struct {
	logs     []string
	maxSize  int
	mu       sync.RWMutex
}

var buffer *LogBuffer

// InitBuffer 初始化日志缓冲区
func InitBuffer(maxSize int) {
	buffer = &LogBuffer{
		logs:    make([]string, 0, maxSize),
		maxSize: maxSize,
	}
}

// AddLog 添加日志到缓冲区
func (lb *LogBuffer) AddLog(log string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.logs = append(lb.logs, log)

	// 保持最大容量
	if len(lb.logs) > lb.maxSize {
		lb.logs = lb.logs[len(lb.logs)-lb.maxSize:]
	}
}

// GetLogs 获取所有日志
func (lb *LogBuffer) GetLogs() []string {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	result := make([]string, len(lb.logs))
	copy(result, lb.logs)
	return result
}

// GetRecentLogs 获取最近 N 条日志
func (lb *LogBuffer) GetRecentLogs(n int) []string {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if n > len(lb.logs) {
		n = len(lb.logs)
	}

	result := make([]string, n)
	copy(result, lb.logs[len(lb.logs)-n:])
	return result
}

// Clear 清空日志
func (lb *LogBuffer) Clear() {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.logs = make([]string, 0, lb.maxSize)
}

// GetBuffer 获取全局缓冲区
func GetBuffer() *LogBuffer {
	return buffer
}
