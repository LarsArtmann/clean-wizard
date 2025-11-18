package adapters

import (
	"os"
	"strconv"
	"time"

	"github.com/caarlos0/env/v6"
)

// EnvironmentConfig holds environment-based configuration
type EnvironmentConfig struct {
	// Application settings
	Debug       bool          `env:"DEBUG" envDefault:"false"`
	Environment string        `env:"ENV" envDefault:"development"`
	LogLevel    string        `env:"LOG_LEVEL" envDefault:"info"`
	Version     string        `env:"VERSION" envDefault:"dev"`

	// Performance settings
	MaxConcurrency int           `env:"MAX_CONCURRENCY" envDefault:"4"`
	Timeout        time.Duration `env:"TIMEOUT" envDefault:"30s"`
	RateLimitRPS   float64       `env:"RATE_LIMIT_RPS" envDefault:"10"`

	// Cache settings
	CacheEnabled         bool          `env:"CACHE_ENABLED" envDefault:"true"`
	CacheTTL           time.Duration `env:"CACHE_TTL" envDefault:"5m"`
	CacheCleanupInterval time.Duration `env:"CACHE_CLEANUP_INTERVAL" envDefault:"10m"`

	// HTTP client settings
	HTTPTimeout         time.Duration `env:"HTTP_TIMEOUT" envDefault:"30s"`
	HTTPRetryCount      int           `env:"HTTP_RETRY_COUNT" envDefault:"3"`
	HTTPRetryWaitTime   time.Duration `env:"HTTP_RETRY_WAIT_TIME" envDefault:"1s"`
	HTTPRetryMaxWait    time.Duration `env:"HTTP_RETRY_MAX_WAIT" envDefault:"10s"`

	// Nix settings
	NixPath              string `env:"NIX_PATH" envDefault:"/nix/var/nix"`
	MaxNixGenerations     int    `env:"MAX_NIX_GENERATIONS" envDefault:"10"`
	DefaultNixGenerations  int    `env:"DEFAULT_NIX_GENERATIONS" envDefault:"3"`
	NixStoreSizeGB        int    `env:"NIX_STORE_SIZE_GB" envDefault:"300"`

	// Disk settings
	MaxDiskUsagePercent int `env:"MAX_DISK_USAGE_PERCENT" envDefault:"50"`
	MinDiskUsagePercent int `env:"MIN_DISK_USAGE_PERCENT" envDefault:"10"`
	RoundingIncrement   int `env:"ROUNDING_INCREMENT" envDefault:"10"`

	// Security settings
	SafeMode           bool `env:"SAFE_MODE" envDefault:"true"`
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
	
	if cfg.MaxNixGenerations <= 0 || cfg.MaxNixGenerations > 1000 {
		return ErrInvalidConfig("max_nix_generations must be between 1 and 1000")
	}
	
	if cfg.MaxDiskUsagePercent < 10 || cfg.MaxDiskUsagePercent > 95 {
		return ErrInvalidConfig("max_disk_usage_percent must be between 10 and 95")
	}
	
	return nil
}

// ToMap converts config to map for logging/debugging
func (cfg *EnvironmentConfig) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"debug":                  cfg.Debug,
		"environment":            cfg.Environment,
		"log_level":              cfg.LogLevel,
		"version":               cfg.Version,
		"max_concurrency":        cfg.MaxConcurrency,
		"timeout":               cfg.Timeout,
		"rate_limit_rps":        cfg.RateLimitRPS,
		"cache_enabled":          cfg.CacheEnabled,
		"cache_ttl":             cfg.CacheTTL,
		"http_timeout":           cfg.HTTPTimeout,
		"nix_path":              cfg.NixPath,
		"max_nix_generations":    cfg.MaxNixGenerations,
		"safe_mode":             cfg.SafeMode,
		"require_confirmation":   cfg.RequireConfirmation,
		"temp_dir":              cfg.TempDir,
		"config_file":           cfg.ConfigFile,
		"state_directory":       cfg.StateDirectory,
	}
}