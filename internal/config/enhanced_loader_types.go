package config

import (
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// RefreshOption represents the force refresh option for config loading.
type RefreshOption string

const (
	RefreshOptionDisabled RefreshOption = "disabled"
	RefreshOptionEnabled  RefreshOption = "enabled"
)

// CacheOption represents the cache enable option for config loading.
type CacheOption string

const (
	CacheOptionDisabled CacheOption = "disabled"
	CacheOptionEnabled  CacheOption = "enabled"
)

// SanitizeOption represents the sanitization enable option.
type SanitizeOption string

const (
	SanitizeOptionDisabled SanitizeOption = "disabled"
	SanitizeOptionEnabled  SanitizeOption = "enabled"
)

// SaveOption represents the force save option for config saving.
type SaveOption string

const (
	SaveOptionDisabled SaveOption = "disabled"
	SaveOptionEnabled  SaveOption = "enabled"
)

// BackupOption represents the backup option for config saving.
type BackupOption string

const (
	BackupOptionDisabled BackupOption = "disabled"
	BackupOptionEnabled  BackupOption = "enabled"
)

// MonitoringOption represents the monitoring option.
type MonitoringOption string

const (
	MonitoringOptionDisabled MonitoringOption = "disabled"
	MonitoringOptionEnabled  MonitoringOption = "enabled"
)

// EnhancedConfigLoader provides comprehensive configuration loading with validation.
type EnhancedConfigLoader struct {
	middleware       *ValidationMiddleware
	validator        *ConfigValidator
	sanitizer        *ConfigSanitizer
	cache            *ConfigCache
	retryPolicy      *RetryPolicy
	enableMonitoring MonitoringOption
}

// ConfigLoadOptions provides options for configuration loading.
type ConfigLoadOptions struct {
	ForceRefresh       RefreshOption              `json:"force_refresh"`
	EnableCache        CacheOption                `json:"enable_cache"`
	EnableSanitization SanitizeOption             `json:"enable_sanitization"`
	ValidationLevel    domain.ValidationLevelType `json:"validation_level"`
	Timeout            time.Duration              `json:"timeout"`
}

// ConfigSaveOptions provides options for configuration saving.
type ConfigSaveOptions struct {
	EnableSanitization SanitizeOption             `json:"enable_sanitization"`
	BackupEnabled      BackupOption               `json:"backup_enabled"`
	ValidationLevel    domain.ValidationLevelType `json:"validation_level"`
	CreateBackup       BackupOption               `json:"create_backup"`
	ForceSave          SaveOption                 `json:"force_save"` // Override validation failures
}

// RetryPolicy defines retry behavior for configuration operations.
type RetryPolicy struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// NewEnhancedConfigLoader creates a new enhanced config loader.
func NewEnhancedConfigLoader(options ...func(*EnhancedConfigLoader)) *EnhancedConfigLoader {
	validator := NewConfigValidator()
	sanitizer := NewConfigSanitizer()
	middleware := NewValidationMiddleware()

	ecl := &EnhancedConfigLoader{
		middleware:       middleware,
		validator:        validator,
		sanitizer:        sanitizer,
		cache:            NewConfigCache(30 * time.Minute),
		retryPolicy:      getDefaultRetryPolicy(),
		enableMonitoring: MonitoringOptionDisabled,
	}

	for _, option := range options {
		option(ecl)
	}

	return ecl
}

// WithMonitoring enables monitoring output.
func WithMonitoring(enabled bool) func(*EnhancedConfigLoader) {
	return func(ecl *EnhancedConfigLoader) {
		if enabled {
			ecl.enableMonitoring = MonitoringOptionEnabled
		} else {
			ecl.enableMonitoring = MonitoringOptionDisabled
		}
	}
}

// WithRetryPolicy sets custom retry policy.
func WithRetryPolicy(policy *RetryPolicy) func(*EnhancedConfigLoader) {
	return func(ecl *EnhancedConfigLoader) {
		// Guard against nil policy to preserve defaults
		if policy == nil {
			return
		}
		ecl.retryPolicy = policy
	}
}
