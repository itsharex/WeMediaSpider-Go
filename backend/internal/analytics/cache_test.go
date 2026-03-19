package analytics

import (
	"testing"
	"time"

	"WeMediaSpider/backend/internal/models"
)

func TestAnalyticsCache_SetAndGet(t *testing.T) {
	cache := &AnalyticsCache{
		data: make(map[string]*cacheEntry),
		ttl:  30 * time.Minute,
	}

	data := &models.AnalyticsData{
		CachedAt: "2026-01-01",
	}
	cache.Set("key1", data)

	got := cache.Get("key1")
	if got == nil {
		t.Fatal("Get returned nil for existing key")
	}
	if got.CachedAt != "2026-01-01" {
		t.Errorf("got CachedAt=%q, want %q", got.CachedAt, "2026-01-01")
	}
}

func TestAnalyticsCache_MissingKey(t *testing.T) {
	cache := &AnalyticsCache{
		data: make(map[string]*cacheEntry),
		ttl:  30 * time.Minute,
	}

	got := cache.Get("nonexistent")
	if got != nil {
		t.Error("Get should return nil for missing key")
	}
}

func TestAnalyticsCache_Expiry(t *testing.T) {
	cache := &AnalyticsCache{
		data: make(map[string]*cacheEntry),
		ttl:  time.Millisecond,
	}

	cache.Set("key", &models.AnalyticsData{CachedAt: "x"})
	time.Sleep(5 * time.Millisecond)

	got := cache.Get("key")
	if got != nil {
		t.Error("Get should return nil for expired entry")
	}
}

func TestAnalyticsCache_Clear(t *testing.T) {
	cache := &AnalyticsCache{
		data: make(map[string]*cacheEntry),
		ttl:  30 * time.Minute,
	}

	cache.Set("a", &models.AnalyticsData{})
	cache.Set("b", &models.AnalyticsData{})
	cache.Clear()

	if cache.Get("a") != nil || cache.Get("b") != nil {
		t.Error("Clear should remove all entries")
	}
}

func TestAnalyticsCache_Overwrite(t *testing.T) {
	cache := &AnalyticsCache{
		data: make(map[string]*cacheEntry),
		ttl:  30 * time.Minute,
	}

	cache.Set("key", &models.AnalyticsData{CachedAt: "first"})
	cache.Set("key", &models.AnalyticsData{CachedAt: "second"})

	got := cache.Get("key")
	if got == nil {
		t.Fatal("Get returned nil")
	}
	if got.CachedAt != "second" {
		t.Errorf("got %q, want %q", got.CachedAt, "second")
	}
}
