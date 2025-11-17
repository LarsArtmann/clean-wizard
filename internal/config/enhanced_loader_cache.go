package config

import (
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// ConfigCache provides configuration caching with TTL
type ConfigCache struct {
	config    *domain.Config
	loadedAt  time.Time
	ttl       time.Duration
	validator *ConfigValidator
}

// NewConfigCache creates a new configuration cache with specified TTL
func NewConfigCache(ttl time.Duration) *ConfigCache {
	return &ConfigCache{
		ttl:       ttl,
		validator: NewConfigValidator(),
	}
}

// Get retrieves cached configuration if valid and not expired
func (cc *ConfigCache) Get() *domain.Config {
	if cc.config == nil || time.Since(cc.loadedAt) > cc.ttl {
		return nil
	}

	// Validate cached config
	result := cc.validator.ValidateConfig(cc.config)
	if !result.IsValid {
		cc.config = nil // Invalidate cache
		return nil
	}

	return cc.config
}

// Set stores configuration in cache with current timestamp
func (cc *ConfigCache) Set(config *domain.Config) {
	cc.config = config
	cc.loadedAt = time.Now()
}
