package adapters

import (
	"os"
	"strconv"
	"time"

	"github.com/caarlos0/env/v6"
)

// Constants for environment validation bounds
const (
	// MaxDiskUsage validation
	MinMaxDiskUsagePercent = 10
	MaxMaxDiskUsagePercent = 95

	// MinDiskUsage validation
	MinDiskUsagePercent = 0
	MaxDiskUsagePercent = 95

	// RoundingIncrement validation
	MinRoundingIncrement = 1
	MaxRoundingIncrement = 50

	// NixGenerations bounds
	MinNixGenerations = 1
	MaxNixGenerations = 1000
)

// EnvironmentConfig holds environment-based configuration
type EnvironmentConfig struct {
	// Application settings
	Debug       bool   `env:"DEBUG" envDefault:"false"`
	Environment string `env:"ENV" envDefault:"development"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
	Version     string `env:"VERSION" envDefault:"dev"`

	// Performance settings
	MaxConcurrency int           `env:"MAX_CONCURRENCY" envDefault:"4"`
	Timeout        time.Duration `env:"TIMEOUT" envDefault:"30s"`
	RateLimitRPS   float64       `env:"RATE_LIMIT_RPS" envDefault:"10"`

	// Cache settings
	CacheEnabled         bool          `env:"CACHE_ENABLED" envDefault:"true"`
	CacheTTL             time.Duration `env:"CACHE_TTL" envDefault:"5m"`
	CacheCleanupInterval time.Duration `env:"CACHE_CLEANUP_INTERVAL" envDefault:"10m"`

	// HTTP client settings
	HTTPTimeout       time.Duration `env:"HTTP_TIMEOUT" envDefault:"30s"`
	HTTPRetryCount    int           `env:"HTTP_RETRY_COUNT" envDefault:"3"`
	HTTPRetryWaitTime time.Duration `env:"HTTP_RETRY_WAIT_TIME" envDefault:"1s"`
	HTTPRetryMaxWait  time.Duration `env:"HTTP_RETRY_MAX_WAIT" envDefault:"10s"`

	// Nix settings
	NixPath               string `env:"NIX_PATH" envDefault:"/nix/var/nix"`
	MaxNixGenerations     int    `env:"MAX_NIX_GENERATIONS" envDefault:"10"`
	DefaultNixGenerations int    `env:"DEFAULT_NIX_GENERATIONS" envDefault:"3"`
	NixStoreSizeGB        int    `env:"NIX_STORE_SIZE_GB" envDefault:"300"`

	// Disk settings
	MaxDiskUsagePercent int `env:"MAX_DISK_USAGE_PERCENT" envDefault:"50"`
	MinDiskUsagePercent int `env:"MIN_DISK_USAGE_PERCENT" envDefault:"10"`
	RoundingIncrement   int `env:"ROUNDING_INCREMENT" envDefault:"10"`

	// Security settings
	SafeMode            bool `env:"SAFE_MODE" envDefault:"true"`
	RequireConfirmation bool `env:"REQUIRE_CONFIRMATION" envDefault:"true"`

	// Filesystem settings
	TempDir        string `env:"TEMP_DIR" envDefault:"/tmp"`
	ConfigFile     string `env:"CONFIG_FILE" envDefault:"clean-wizard.yaml"`
	StateDirectory string `env:"STATE_DIRECTORY" envDefault:"~/.clean-wizard"`

	// Monitoring settings
	MetricsEnabled  bool   `env:"METRICS_ENABLED" envDefault:"false"`
	MetricsPort     int    `env:"METRICS_PORT" envDefault:"8080"`
	MetricsPath     string `env:"METRICS_PATH" envDefault:"/metrics"`
	TracingEnabled  bool   `env:"TRACING_ENABLED" envDefault:"false"`
	TracingEndpoint string `env:"TRACING_ENDPOINT" envDefault:""`
}

// LoadEnvironmentConfig loads configuration from environment variables
func LoadEnvironmentConfig() (*EnvironmentConfig, error) {
	cfg := &EnvironmentConfig{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	if err := cfg.ValidateEnvironmentConfig(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadEnvironmentConfigWithPrefix loads configuration with custom prefix
func LoadEnvironmentConfigWithPrefix(prefix string) (*EnvironmentConfig, error) {
	cfg := &EnvironmentConfig{}

	if err := env.Parse(cfg, env.Options{
		Prefix: prefix,
	}); err != nil {
		return nil, err
	}

	if err := cfg.ValidateEnvironmentConfig(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// GetEnvWithDefault returns environment variable with default value
func GetEnvWithDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// GetEnvBool returns boolean environment variable with default
func GetEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// GetEnvInt returns integer environment variable with default
func GetEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// GetEnvDuration returns duration environment variable with default
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// ValidateEnvironmentConfig validates the loaded configuration
func (cfg *EnvironmentConfig) ValidateEnvironmentConfig() error {
	if cfg.MaxConcurrency <= 0 {
		return ErrInvalidConfig("max_concurrency must be positive")
	}

	if cfg.Timeout <= 0 {
		return ErrInvalidConfig("timeout must be positive")
	}

	if cfg.RateLimitRPS <= 0 {
		return ErrInvalidConfig("rate_limit_rps must be positive")
	}

	if cfg.MaxNixGenerations < MinNixGenerations || cfg.MaxNixGenerations > MaxNixGenerations {
		return ErrInvalidConfig("max_nix_generations must be between %d and %d", MinNixGenerations, MaxNixGenerations)
	}

	if cfg.MaxDiskUsagePercent < MinMaxDiskUsagePercent || cfg.MaxDiskUsagePercent > MaxMaxDiskUsagePercent {
		return ErrInvalidConfig("max_disk_usage_percent must be between %d and %d", MinMaxDiskUsagePercent, MaxMaxDiskUsagePercent)
	}

	// NEW: Validate MinDiskUsagePercent
	if cfg.MinDiskUsagePercent < MinDiskUsagePercent || cfg.MinDiskUsagePercent > MaxDiskUsagePercent {
		return ErrInvalidConfig("min_disk_usage_percent must be between %d and %d", MinDiskUsagePercent, MaxDiskUsagePercent)
	}

	// NEW: Ensure MinDiskUsagePercent doesn't exceed MaxDiskUsagePercent
	if cfg.MinDiskUsagePercent > cfg.MaxDiskUsagePercent {
		return ErrInvalidConfig("min_disk_usage_percent (%d) cannot exceed max_disk_usage_percent (%d)",
			cfg.MinDiskUsagePercent, cfg.MaxDiskUsagePercent)
	}

	// NEW: Validate RoundingIncrement
	if cfg.RoundingIncrement < MinRoundingIncrement || cfg.RoundingIncrement > MaxRoundingIncrement {
		return ErrInvalidConfig("rounding_increment must be between %d and %d", MinRoundingIncrement, MaxRoundingIncrement)
	}

	return nil
}

// EnvironmentConfigView provides strongly-typed config representation for logging/debugging
type EnvironmentConfigView struct {
	Debug               bool          `json:"debug"`
	Environment         string        `json:"environment"`
	LogLevel            string        `json:"log_level"`
	Version             string        `json:"version"`
	MaxConcurrency      int           `json:"max_concurrency"`
	Timeout             time.Duration `json:"timeout"`
	RateLimitRPS        float64       `json:"rate_limit_rps"`
	CacheEnabled        bool          `json:"cache_enabled"`
	CacheTTL            time.Duration `json:"cache_ttl"`
	HTTPTimeout         time.Duration `json:"http_timeout"`
	NixPath             string        `json:"nix_path"`
	MaxNixGenerations   int           `json:"max_nix_generations"`
	SafeMode            bool          `json:"safe_mode"`
	RequireConfirmation bool          `json:"require_confirmation"`
	TempDir             string        `json:"temp_dir"`
	ConfigFile          string        `json:"config_file"`
	StateDirectory      string        `json:"state_directory"`
}

// ToView converts config to strongly-typed view for logging/debugging
func (cfg *EnvironmentConfig) ToView() EnvironmentConfigView {
	return EnvironmentConfigView{
		Debug:               cfg.Debug,
		Environment:         cfg.Environment,
		LogLevel:            cfg.LogLevel,
		Version:             cfg.Version,
		MaxConcurrency:      cfg.MaxConcurrency,
		Timeout:             cfg.Timeout,
		RateLimitRPS:        cfg.RateLimitRPS,
		CacheEnabled:        cfg.CacheEnabled,
		CacheTTL:            cfg.CacheTTL,
		HTTPTimeout:         cfg.HTTPTimeout,
		NixPath:             cfg.NixPath,
		MaxNixGenerations:   cfg.MaxNixGenerations,
		SafeMode:            cfg.SafeMode,
		RequireConfirmation: cfg.RequireConfirmation,
		TempDir:             cfg.TempDir,
		ConfigFile:          cfg.ConfigFile,
		StateDirectory:      cfg.StateDirectory,
	}
}

// ToMap converts config to map for legacy compatibility (deprecated)
// TYPE-SAFE-EXEMPT: Legacy compatibility method using map[string]any intentionally
func (cfg *EnvironmentConfig) ToMap() map[string]any {
	view := cfg.ToView()
	return map[string]any{
		"debug":                view.Debug,
		"environment":          view.Environment,
		"log_level":            view.LogLevel,
		"version":              view.Version,
		"max_concurrency":      view.MaxConcurrency,
		"timeout":              view.Timeout.String(),
		"rate_limit_rps":       view.RateLimitRPS,
		"cache_enabled":        view.CacheEnabled,
		"cache_ttl":            view.CacheTTL.String(),
		"http_timeout":         view.HTTPTimeout.String(),
		"nix_path":             view.NixPath,
		"max_nix_generations":  view.MaxNixGenerations,
		"safe_mode":            view.SafeMode,
		"require_confirmation": view.RequireConfirmation,
		"temp_dir":             view.TempDir,
		"config_file":          view.ConfigFile,
		"state_directory":      view.StateDirectory,
	}
}
