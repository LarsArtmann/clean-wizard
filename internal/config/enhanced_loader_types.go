package config

import (
	"time"
)

// ValidationLevel defines validation strictness (backward compatibility)
//
// IMPORTANT: These integer values MUST remain in sync with domain.ValidationLevelType
// See: internal/domain/type_safe_enums.go:116-123
// When changing numeric values, update BOTH locations to prevent drift.
// Consider aliasing the domain type directly to avoid duplication in the future.
type ValidationLevel int

const (
	ValidationLevelNone          ValidationLevel = 0
	ValidationLevelBasic         ValidationLevel = 1
	ValidationLevelComprehensive ValidationLevel = 2
	ValidationLevelStrict        ValidationLevel = 3
)

// String returns string representation
func (vl ValidationLevel) String() string {
	switch vl {
	case ValidationLevelNone:
		return "NONE"
	case ValidationLevelBasic:
		return "BASIC"
	case ValidationLevelComprehensive:
		return "COMPREHENSIVE"
	case ValidationLevelStrict:
		return "STRICT"
	default:
		return "UNKNOWN"
	}
}

// EnhancedConfigLoader provides comprehensive configuration loading with validation
type EnhancedConfigLoader struct {
	middleware       *ValidationMiddleware
	validator        *ConfigValidator
	sanitizer        *ConfigSanitizer
	cache            *ConfigCache
	retryPolicy      *RetryPolicy
	enableMonitoring bool
}

// ConfigLoadOptions provides options for configuration loading
type ConfigLoadOptions struct {
	ForceRefresh       bool            `json:"force_refresh"`
	EnableCache        bool            `json:"enable_cache"`
	EnableSanitization bool            `json:"enable_sanitization"`
	ValidationLevel    ValidationLevel `json:"validation_level"`
	Timeout            time.Duration   `json:"timeout"`
}

// ConfigSaveOptions provides options for configuration saving
type ConfigSaveOptions struct {
	EnableSanitization bool            `json:"enable_sanitization"`
	CreateBackup       bool            `json:"create_backup"` // Whether to create a backup before saving
	ValidationLevel    ValidationLevel `json:"validation_level"`
	ForceSave          bool            `json:"force_save"` // Override validation failures
}

// RetryPolicy defines retry behavior for configuration operations
type RetryPolicy struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// NewEnhancedConfigLoader creates a new enhanced config loader
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
		enableMonitoring: false,
	}

	for _, option := range options {
		option(ecl)
	}

	return ecl
}

// WithMonitoring enables monitoring output
func WithMonitoring(enabled bool) func(*EnhancedConfigLoader) {
	return func(ecl *EnhancedConfigLoader) {
		ecl.enableMonitoring = enabled
	}
}

// WithRetryPolicy sets custom retry policy
func WithRetryPolicy(policy *RetryPolicy) func(*EnhancedConfigLoader) {
	return func(ecl *EnhancedConfigLoader) {
		// Guard against nil policy to preserve defaults
		if policy == nil {
			return
		}
		ecl.retryPolicy = policy
	}
}
