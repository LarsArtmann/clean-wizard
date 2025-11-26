package config

import (
	"context"
	"time"

	domainconfig "github.com/LarsArtmann/clean-wizard/internal/domain/config"
)

// ConfigLoader provides application-level configuration loading services
type ConfigLoader struct {
	defaultTimeout time.Duration
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{
		defaultTimeout: 30 * time.Second,
	}
}

// Load loads configuration using the application config package
func (cl *ConfigLoader) Load(ctx context.Context) (*domainconfig.Config, error) {
	return LoadWithContext(ctx)
}

// Save saves configuration using the application config package
func (cl *ConfigLoader) Save(ctx context.Context, cfg *domainconfig.Config) error {
	return Save(cfg)
}
