package adapters

import (
	"time"

	"github.com/maypok86/otter/v2"
)

// Cache configuration constants.
const (
	// DefaultCacheMaximumSize is the default maximum number of items in the cache.
	DefaultCacheMaximumSize = 10_000
	// DefaultCacheExpiry is the default expiration time for cache entries.
	DefaultCacheExpiry = 24 * time.Hour
)

// cacheItem wraps a value with its expiration time.
type cacheItem struct {
	value      any
	expiration time.Time
}

// isExpired checks if the item has expired.
func (ci *cacheItem) isExpired() bool {
	return time.Now().After(ci.expiration)
}

// CacheManager provides in-memory caching functionality.
type CacheManager struct {
	cache *otter.Cache[string, *cacheItem]
}

// NewCacheManager creates a new cache manager. Entries expire lazily on read
// after DefaultCacheExpiry (otter's expiry calculator); there is no background
// cleanup loop.
func NewCacheManager() *CacheManager {
	cache := otter.Must(&otter.Options[string, *cacheItem]{
		MaximumSize: DefaultCacheMaximumSize,
		ExpiryCalculator: otter.ExpiryWriting[string, *cacheItem](DefaultCacheExpiry),
	})

	return &CacheManager{cache: cache}
}

// Set stores a value in cache with expiration.
func (cm *CacheManager) Set(key string, value any, expiration time.Duration) {
	item := &cacheItem{
		value:      value,
		expiration: time.Now().Add(expiration),
	}
	cm.cache.Set(key, item)
}

// Get retrieves a value from cache.
func (cm *CacheManager) Get(key string) (any, bool) {
	item, ok := cm.cache.GetIfPresent(key)
	if !ok {
		return nil, false
	}

	// Check if expired
	if item.isExpired() {
		cm.cache.Invalidate(key)

		return nil, false
	}

	return item.value, true
}

// GetWithExpiration retrieves a value with its expiration time.
func (cm *CacheManager) GetWithExpiration(key string) (any, time.Time, bool) {
	item, ok := cm.cache.GetIfPresent(key)
	if !ok {
		return nil, time.Time{}, false
	}

	// Check if expired
	if item.isExpired() {
		cm.cache.Invalidate(key)

		return nil, time.Time{}, false
	}

	return item.value, item.expiration, true
}

// Delete removes an item from cache.
func (cm *CacheManager) Delete(key string) {
	cm.cache.Invalidate(key)
}

// Clear removes all items from cache.
func (cm *CacheManager) Clear() {
	cm.cache.InvalidateAll()
}

// ItemCount returns the number of items in cache.
func (cm *CacheManager) ItemCount() int {
	return cm.cache.EstimatedSize()
}

// Shutdown invalidates every cached entry when the DI container shuts down.
func (cm *CacheManager) Shutdown() error {
	cm.cache.InvalidateAll()

	return nil
}

// CacheStats provides cache statistics.
type CacheStats struct {
	Items     int     `json:"items"`
	HitCount  uint64  `json:"hitCount"`
	MissCount uint64  `json:"missCount"`
	HitRate   float64 `json:"hitRate"`
}

// Stats returns cache performance statistics.
func (cm *CacheManager) Stats() CacheStats {
	stats := cm.cache.Stats()
	estimatedSize := cm.cache.EstimatedSize()

	hitRate := float64(0)
	if total := stats.Hits + stats.Misses; total > 0 {
		hitRate = float64(stats.Hits) / float64(total)
	}

	return CacheStats{
		Items:     estimatedSize,
		HitCount:  stats.Hits,
		MissCount: stats.Misses,
		HitRate:   hitRate,
	}
}
