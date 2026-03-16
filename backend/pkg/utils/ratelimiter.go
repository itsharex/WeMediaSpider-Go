package utils

import (
	"math/rand"
	"sync"
	"time"

	"WeMediaSpider/backend/pkg/timeutil"
)

// RateLimiter 智能请求频率限制器（基于令牌桶算法，支持高并发）
type RateLimiter struct {
	mu              sync.Mutex
	tokens          int           // 当前可用令牌数
	maxTokens       int           // 最大令牌数
	refillRate      time.Duration // 令牌补充速率
	lastRefill      time.Time     // 上次补充时间
	failureCount    int           // 连续失败次数
	lastFailureTime time.Time
	adaptiveMode    bool // 自适应模式
}

// NewRateLimiter 创建智能频率限制器
// minInterval: 最小请求间隔（用于计算令牌补充速率）
// maxInterval: 最大请求间隔（未使用，保留兼容性）
// maxRequests: 令牌桶容量（同时允许的最大并发请求数）
func NewRateLimiter(minInterval, maxInterval time.Duration, maxRequests int) *RateLimiter {
	return &RateLimiter{
		tokens:       maxRequests,
		maxTokens:    maxRequests,
		refillRate:   minInterval,
		lastRefill:   timeutil.Now(),
		failureCount: 0,
		adaptiveMode: true,
	}
}

// Wait 等待直到可以发送下一个请求（非阻塞式令牌获取）
func (rl *RateLimiter) Wait() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := timeutil.Now()

	// 补充令牌
	rl.refillTokens(now)

	// 如果没有可用令牌，等待
	for rl.tokens <= 0 {
		// 计算需要等待的时间
		waitTime := rl.refillRate

		// 自适应延迟：根据失败次数动态调整
		if rl.adaptiveMode && rl.failureCount > 0 {
			multiplier := 1.0 + float64(rl.failureCount)*0.3
			if multiplier > 2.0 {
				multiplier = 2.0
			}
			waitTime = time.Duration(float64(waitTime) * multiplier)
		}

		// 释放锁，等待，然后重新获取锁
		rl.mu.Unlock()
		time.Sleep(waitTime)
		rl.mu.Lock()

		// 重新补充令牌
		now = timeutil.Now()
		rl.refillTokens(now)
	}

	// 消耗一个令牌
	rl.tokens--
}

// refillTokens 补充令牌（内部方法，调用前必须持有锁）
func (rl *RateLimiter) refillTokens(now time.Time) {
	elapsed := now.Sub(rl.lastRefill)

	// 计算应该补充的令牌数
	tokensToAdd := int(elapsed / rl.refillRate)

	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastRefill = now
	}
}

// RecordSuccess 记录成功请求（降低失败计数）
func (rl *RateLimiter) RecordSuccess() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.failureCount > 0 {
		rl.failureCount--
	}
}

// RecordFailure 记录失败请求（增加失败计数）
func (rl *RateLimiter) RecordFailure() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.failureCount++
	rl.lastFailureTime = timeutil.Now()
	if rl.failureCount > 10 {
		rl.failureCount = 10 // 最多记录10次
	}
}

// GetCurrentDelay 获取当前延迟时间（用于日志）
func (rl *RateLimiter) GetCurrentDelay() time.Duration {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.failureCount > 0 {
		multiplier := 1.0 + float64(rl.failureCount)*0.3
		if multiplier > 2.0 {
			multiplier = 2.0
		}
		return time.Duration(float64(rl.refillRate) * multiplier)
	}
	return rl.refillRate
}

// GetRandomDelay 获取随机延迟时间
func GetRandomDelay(min, max time.Duration) time.Duration {
	if max <= min {
		return min
	}
	return min + time.Duration(rand.Int63n(int64(max-min)))
}

// GetExponentialBackoff 获取指数退避延迟
func GetExponentialBackoff(attempt int, baseDelay time.Duration) time.Duration {
	if attempt <= 0 {
		return baseDelay
	}
	// 2^attempt * baseDelay，最大不超过2分钟（降低最大延迟）
	delay := baseDelay * time.Duration(1<<uint(attempt))
	maxDelay := 2 * time.Minute
	if delay > maxDelay {
		delay = maxDelay
	}
	// 添加随机抖动（±20%）
	jitter := time.Duration(rand.Int63n(int64(delay) / 5))
	if rand.Intn(2) == 0 {
		delay += jitter
	} else {
		delay -= jitter
	}
	return delay
}
