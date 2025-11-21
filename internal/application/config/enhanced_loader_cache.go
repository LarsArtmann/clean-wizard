package config

import (
	"sync"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// ConfigCache provides configuration caching with TTL
type ConfigCache struct {
	mu        sync.RWMutex
	config    *config.Config
	loadedAt  time.Time
	ttl       time.Duration
	validator *ConfigValidator
}

// NewConfigCache creates a new configuration cache with specified TTL
func NewConfigCache(ttl time.Duration) *config.ConfigCache {
	return &ConfigCache{
		ttl:       ttl,
		validator: NewConfigValidator(),
	}
}

// Get retrieves cached configuration if valid and not expired
func (cc *config.ConfigCache) Get() *config.Config {
	cc.mu.RLock()
	if cc.config == nil || time.Since(cc.loadedAt) > cc.ttl {
		cc.mu.RUnlock()
		return nil
	}

	// Validate cached config
	validator := cc.validator
	config := cc.config
	cc.mu.RUnlock()

	result := validator.ValidateConfig(config)
	if !result.IsValid {
		cc.mu.Lock()
		cc.config = nil // Invalidate cache
		cc.mu.Unlock()
		return nil
	}

	return config
}

// Set stores configuration in cache with current timestamp
func (cc *config.ConfigCache) Set(config *config.Config) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.config = config
	cc.loadedAt = time.Now()
}
