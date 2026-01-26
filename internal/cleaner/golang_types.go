package cleaner

import (
	"fmt"
)

// GoCacheType represents a type-safe enum for Go cache flags.
// It uses bit flags so multiple cache types can be combined.
type GoCacheType uint16

const (
	// GoCacheNone represents no cache types - invalid state
	GoCacheNone GoCacheType = 0
	// GoCacheGOCACHE represents the main Go cache (GOCACHE)
	GoCacheGOCACHE GoCacheType = 1 << iota
	// GoCacheTestCache represents the Go test cache (GOTESTCACHE)
	GoCacheTestCache
	// GoCacheModCache represents the Go module cache (GOMODCACHE)
	GoCacheModCache
	// GoCacheBuildCache represents the Go build cache folders (go-build*)
	GoCacheBuildCache
	// GoCacheLintCache represents the lint cache (e.g., golangci-lint)
	GoCacheLintCache
)

// IsValid checks if at least one cache type is enabled.
func (gt GoCacheType) IsValid() bool {
	return gt != GoCacheNone
}

// Has checks if the given cache type is enabled.
func (gt GoCacheType) Has(cacheType GoCacheType) bool {
	return gt&cacheType != 0
}

// Count returns the number of enabled cache types.
func (gt GoCacheType) Count() int {
	count := 0
	for gt != 0 {
		count++
		gt &= gt - 1 // Clear the lowest set bit
	}
	return count
}

// EnabledTypes returns a slice of all enabled cache types.
func (gt GoCacheType) EnabledTypes() []GoCacheType {
	var types []GoCacheType
	allTypes := []GoCacheType{
		GoCacheGOCACHE,
		GoCacheTestCache,
		GoCacheModCache,
		GoCacheBuildCache,
		GoCacheLintCache,
	}
	for _, t := range allTypes {
		if gt.Has(t) {
			types = append(types, t)
		}
	}
	return types
}

// String returns a human-readable string for the cache type.
func (gt GoCacheType) String() string {
	if gt == GoCacheGOCACHE {
		return "Go Cache (GOCACHE)"
	}
	if gt == GoCacheTestCache {
		return "Go Test Cache (GOTESTCACHE)"
	}
	if gt == GoCacheModCache {
		return "Go Module Cache (GOMODCACHE)"
	}
	if gt == GoCacheBuildCache {
		return "Go Build Cache (go-build)"
	}
	if gt == GoCacheLintCache {
		return "Lint Cache (golangci-lint)"
	}
	if gt.Count() > 1 {
		return fmt.Sprintf("Multiple caches (%d)", gt.Count())
	}
	return "Unknown cache"
}
