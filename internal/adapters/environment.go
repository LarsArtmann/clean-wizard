package adapters

import (
	"os"
	"strconv"
	"time"

	"github.com/caarlos0/env/v6"
)

// AppSettings holds application-level configuration.
type AppSettings struct {
	Debug       bool   `env:"DEBUG"     envDefault:"false"`
	Environment string `env:"ENV"       envDefault:"development"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
	Version     string `env:"VERSION"   envDefault:"dev"`
}

// PerformanceSettings holds performance-related configuration.
type PerformanceSettings struct {
	MaxConcurrency int           `env:"MAX_CONCURRENCY" envDefault:"4"`
	Timeout        time.Duration `env:"TIMEOUT"         envDefault:"30s"`
	RateLimitRPS   float64       `env:"RATE_LIMIT_RPS"  envDefault:"10"`
}

// CacheSettings holds cache-related configuration.
type CacheSettings struct {
	Enabled         bool          `env:"CACHE_ENABLED"          envDefault:"true"`
	TTL             time.Duration `env:"CACHE_TTL"              envDefault:"5m"`
	CleanupInterval time.Duration `env:"CACHE_CLEANUP_INTERVAL" envDefault:"10m"`
}

// HTTPSettings holds HTTP client configuration.
type HTTPSettings struct {
	Timeout       time.Duration `env:"HTTP_TIMEOUT"         envDefault:"30s"`
	RetryCount    int           `env:"HTTP_RETRY_COUNT"     envDefault:"3"`
	RetryWaitTime time.Duration `env:"HTTP_RETRY_WAIT_TIME" envDefault:"1s"`
	RetryMaxWait  time.Duration `env:"HTTP_RETRY_MAX_WAIT"  envDefault:"10s"`
}

// NixSettings holds Nix-specific configuration.
type NixSettings struct {
	Path               string `env:"NIX_PATH"                envDefault:"/nix/var/nix"`
	MaxGenerations     int    `env:"MAX_NIX_GENERATIONS"     envDefault:"10"`
	DefaultGenerations int    `env:"DEFAULT_NIX_GENERATIONS" envDefault:"3"`
	StoreSizeGB        int    `env:"NIX_STORE_SIZE_GB"       envDefault:"300"`
}

// DiskSettings holds disk usage configuration.
type DiskSettings struct {
	MaxUsagePercent   int `env:"MAX_DISK_USAGE_PERCENT" envDefault:"50"`
	MinUsagePercent   int `env:"MIN_DISK_USAGE_PERCENT" envDefault:"10"`
	RoundingIncrement int `env:"ROUNDING_INCREMENT"     envDefault:"10"`
}

// SecuritySettings holds security-related configuration.
type SecuritySettings struct {
	SafeMode            bool `env:"SAFE_MODE"            envDefault:"true"`
	RequireConfirmation bool `env:"REQUIRE_CONFIRMATION" envDefault:"true"`
}

// FilesystemSettings holds filesystem paths configuration.
type FilesystemSettings struct {
	TempDir        string `env:"TEMP_DIR"        envDefault:"/tmp"`
	ConfigFile     string `env:"CONFIG_FILE"     envDefault:"clean-wizard.yaml"`
	StateDirectory string `env:"STATE_DIRECTORY" envDefault:"~/.clean-wizard"`
}

// MonitoringSettings holds monitoring and observability configuration.
type MonitoringSettings struct {
	MetricsEnabled  bool   `env:"METRICS_ENABLED"  envDefault:"false"`
	MetricsPort     int    `env:"METRICS_PORT"     envDefault:"8080"`
	MetricsPath     string `env:"METRICS_PATH"     envDefault:"/metrics"`
	TracingEnabled  bool   `env:"TRACING_ENABLED"  envDefault:"false"`
	TracingEndpoint string `env:"TRACING_ENDPOINT" envDefault:""`
}

// EnvironmentConfig holds environment-based configuration using composition.
type EnvironmentConfig struct {
	App         AppSettings
	Performance PerformanceSettings
	Cache       CacheSettings
	HTTP        HTTPSettings
	Nix         NixSettings
	Disk        DiskSettings
	Security    SecuritySettings
	Filesystem  FilesystemSettings
	Monitoring  MonitoringSettings
}

// LoadEnvironmentConfig loads configuration from environment variables.
func LoadEnvironmentConfig() (*EnvironmentConfig, error) {
	cfg := &EnvironmentConfig{}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadEnvironmentConfigWithPrefix loads configuration with custom prefix.
func LoadEnvironmentConfigWithPrefix(prefix string) (*EnvironmentConfig, error) {
	cfg := &EnvironmentConfig{}

	err := env.Parse(cfg, env.Options{
		Prefix: prefix,
	})
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// GetEnvWithDefault returns environment variable with default value.
func GetEnvWithDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}

// GetEnvBool returns boolean environment variable with default.
func GetEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}

	return defaultValue
}

// GetEnvInt returns integer environment variable with default.
func GetEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}

	return defaultValue
}

// GetEnvDuration returns duration environment variable with default.
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}

	return defaultValue
}

// ValidateEnvironmentConfig validates the loaded configuration.
func (cfg *EnvironmentConfig) ValidateEnvironmentConfig() error {
	if cfg.Performance.MaxConcurrency <= 0 {
		return ErrInvalidConfig("max_concurrency must be positive")
	}

	if cfg.Performance.Timeout <= 0 {
		return ErrInvalidConfig("timeout must be positive")
	}

	if cfg.Performance.RateLimitRPS <= 0 {
		return ErrInvalidConfig("rate_limit_rps must be positive")
	}

	if cfg.Nix.MaxGenerations <= 0 || cfg.Nix.MaxGenerations > 1000 {
		return ErrInvalidConfig("max_nix_generations must be between 1 and 1000")
	}

	if cfg.Disk.MaxUsagePercent < 10 || cfg.Disk.MaxUsagePercent > 95 {
		return ErrInvalidConfig("max_disk_usage_percent must be between 10 and 95")
	}

	return nil
}

// EnvironmentConfigView provides strongly-typed config representation for logging/debugging.
type EnvironmentConfigView struct {
	Debug               bool          `json:"debug"`
	Environment         string        `json:"environment"`
	LogLevel            string        `json:"logLevel"`
	Version             string        `json:"version"`
	MaxConcurrency      int           `json:"maxConcurrency"`
	Timeout             time.Duration `json:"timeout"`
	RateLimitRPS        float64       `json:"rateLimitRps"`
	CacheEnabled        bool          `json:"cacheEnabled"`
	CacheTTL            time.Duration `json:"cacheTtl"`
	HTTPTimeout         time.Duration `json:"httpTimeout"`
	NixPath             string        `json:"nixPath"`
	MaxNixGenerations   int           `json:"maxNixGenerations"`
	SafeMode            bool          `json:"safeMode"`
	RequireConfirmation bool          `json:"requireConfirmation"`
	TempDir             string        `json:"tempDir"`
	ConfigFile          string        `json:"configFile"`
	StateDirectory      string        `json:"stateDirectory"`
}

// ToView converts config to strongly-typed view for logging/debugging.
func (cfg *EnvironmentConfig) ToView() EnvironmentConfigView {
	return EnvironmentConfigView{
		Debug:               cfg.App.Debug,
		Environment:         cfg.App.Environment,
		LogLevel:            cfg.App.LogLevel,
		Version:             cfg.App.Version,
		MaxConcurrency:      cfg.Performance.MaxConcurrency,
		Timeout:             cfg.Performance.Timeout,
		RateLimitRPS:        cfg.Performance.RateLimitRPS,
		CacheEnabled:        cfg.Cache.Enabled,
		CacheTTL:            cfg.Cache.TTL,
		HTTPTimeout:         cfg.HTTP.Timeout,
		NixPath:             cfg.Nix.Path,
		MaxNixGenerations:   cfg.Nix.MaxGenerations,
		SafeMode:            cfg.Security.SafeMode,
		RequireConfirmation: cfg.Security.RequireConfirmation,
		TempDir:             cfg.Filesystem.TempDir,
		ConfigFile:          cfg.Filesystem.ConfigFile,
		StateDirectory:      cfg.Filesystem.StateDirectory,
	}
}

// ToMap converts config to map for legacy compatibility (deprecated).
func (cfg *EnvironmentConfig) ToMap() map[string]any {
	view := cfg.ToView()

	return map[string]any{
		"debug":               view.Debug,
		"environment":         view.Environment,
		"logLevel":            view.LogLevel,
		"version":             view.Version,
		"maxConcurrency":      view.MaxConcurrency,
		"timeout":             view.Timeout.String(),
		"rateLimitRps":        view.RateLimitRPS,
		"cacheEnabled":        view.CacheEnabled,
		"cacheTtl":            view.CacheTTL.String(),
		"httpTimeout":         view.HTTPTimeout.String(),
		"nixPath":             view.NixPath,
		"maxNixGenerations":   view.MaxNixGenerations,
		"safeMode":            view.SafeMode,
		"requireConfirmation": view.RequireConfirmation,
		"tempDir":             view.TempDir,
		"configFile":          view.ConfigFile,
		"stateDirectory":      view.StateDirectory,
	}
}
