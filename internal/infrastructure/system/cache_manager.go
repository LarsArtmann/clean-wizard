package system

import (
	"time"

	"github.com/patrickmn/go-cache"
)

// CacheManager provides in-memory caching functionality
type CacheManager struct {
	cache *cache.Cache
}

// NewCacheManager creates a new cache manager
// defaultExpiration: default TTL for cache entries
// cleanupInterval: interval for cleaning expired entries
func NewCacheManager(defaultExpiration, cleanupInterval time.Duration) *CacheManager {
	return &CacheManager{
		cache: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Set stores a value in cache with expiration
func (cm *CacheManager) Set(key string, value any, expiration time.Duration) {
	cm.cache.Set(key, value, expiration)
}

// Get retrieves a value from cache
func (cm *CacheManager) Get(key string) (any, bool) {
	return cm.cache.Get(key)
}

// GetWithExpiration retrieves a value with its expiration time
func (cm *CacheManager) GetWithExpiration(key string) (any, time.Time, bool) {
	return cm.cache.GetWithExpiration(key)
}

// Delete removes an item from cache
func (cm *CacheManager) Delete(key string) {
	cm.cache.Delete(key)
}

// Clear removes all items from cache
func (cm *CacheManager) Clear() {
	cm.cache.Flush()
}

// ItemCount returns the number of items in cache
func (cm *CacheManager) ItemCount() int {
	return cm.cache.ItemCount()
}

// CacheStats provides cache statistics
type CacheStats struct {
	Items     int     `json:"items"`
	HitCount  uint64  `json:"hit_count"`
	MissCount uint64  `json:"miss_count"`
	HitRate   float64 `json:"hit_rate"`
}

// Stats returns cache performance statistics
func (cm *CacheManager) Stats() CacheStats {
	// Note: go-cache doesn't provide hit/miss counts by default
	// In production, we'd wrap with instrumentation
	return CacheStats{
		Items: cm.cache.ItemCount(),
	}
}

// Keys returns all keys in the cache
func (cm *CacheManager) Keys() []string {
	items := cm.cache.Items()
	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}
	return keys
}

// SetDefault sets the default expiration for new items
func (cm *CacheManager) SetDefault(expiration time.Duration) {
	// go-cache doesn't support setting default after creation
	// This is a no-op but maintains interface compatibility
}

// FlushExpired removes all expired items
func (cm *CacheManager) FlushExpired() {
	for key, item := range cm.cache.Items() {
		if item.Expired() {
			cm.cache.Delete(key)
		}
	}
}
