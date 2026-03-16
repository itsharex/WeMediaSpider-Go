package analytics

import (
	"sync"
	"time"

	"WeMediaSpider/backend/internal/models"
	"WeMediaSpider/backend/pkg/timeutil"
)

// AnalyticsCache 分析结果缓存
type AnalyticsCache struct {
	data map[string]*cacheEntry
	mu   sync.RWMutex
	ttl  time.Duration
}

type cacheEntry struct {
	data      *models.AnalyticsData
	expiresAt time.Time
}

// NewAnalyticsCache 创建缓存实例
func NewAnalyticsCache(ttl time.Duration) *AnalyticsCache {
	cache := &AnalyticsCache{
		data: make(map[string]*cacheEntry),
		ttl:  ttl,
	}

	// 启动清理协程
	go cache.cleanupLoop()

	return cache
}

// Get 获取缓存数据
func (c *AnalyticsCache) Get(key string) *models.AnalyticsData {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return nil
	}

	if timeutil.Now().After(entry.expiresAt) {
		return nil
	}

	return entry.data
}

// Set 设置缓存数据
func (c *AnalyticsCache) Set(key string, data *models.AnalyticsData) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = &cacheEntry{
		data:      data,
		expiresAt: timeutil.Now().Add(c.ttl),
	}
}

// Clear 清除所有缓存
func (c *AnalyticsCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*cacheEntry)
}

// cleanupLoop 定期清理过期缓存
func (c *AnalyticsCache) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup 清理过期缓存
func (c *AnalyticsCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := timeutil.Now()
	for key, entry := range c.data {
		if now.After(entry.expiresAt) {
			delete(c.data, key)
		}
	}
}
