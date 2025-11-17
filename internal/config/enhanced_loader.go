package config

import (
	"context"
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	pkgerrors "github.com/LarsArtmann/clean-wizard/internal/pkg/errors"
)

// EnhancedConfigLoader provides comprehensive configuration loading with validation
type EnhancedConfigLoader struct {
	middleware       *ValidationMiddleware
	validator        *ConfigValidator
	sanitizer        *ConfigSanitizer
	cache            *ConfigCache
	retryPolicy      *RetryPolicy
	enableMonitoring bool
}

// RetryPolicy defines retry behavior for configuration operations
type RetryPolicy struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
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
	BackupEnabled      bool            `json:"backup_enabled"`
	ValidationLevel    ValidationLevel `json:"validation_level"`
	CreateBackup       bool            `json:"create_backup"`
}

// ValidationLevel defines validation strictness
type ValidationLevel int

const (
	ValidationLevelNone ValidationLevel = iota
	ValidationLevelBasic
	ValidationLevelComprehensive
	ValidationLevelStrict
)

// String returns string representation of ValidationLevel
func (vl ValidationLevel) String() string {
	switch vl {
	case ValidationLevelNone:
		return "None"
	case ValidationLevelBasic:
		return "Basic"
	case ValidationLevelComprehensive:
		return "Comprehensive"
	case ValidationLevelStrict:
		return "Strict"
	default:
		return "Unknown"
	}
}

// NewEnhancedConfigLoader creates an enhanced configuration loader
func NewEnhancedConfigLoader() *EnhancedConfigLoader {
	return &EnhancedConfigLoader{
		middleware:       NewValidationMiddleware(),
		validator:        NewConfigValidator(),
		sanitizer:        NewConfigSanitizer(),
		cache:            NewConfigCache(30 * time.Minute),
		retryPolicy:      getDefaultRetryPolicy(),
		enableMonitoring: false,
	}
}

// NewEnhancedConfigLoaderWithOptions creates loader with custom options
func NewEnhancedConfigLoaderWithOptions(middleware *ValidationMiddleware, validator *ConfigValidator, sanitizer *ConfigSanitizer) *EnhancedConfigLoader {
	return &EnhancedConfigLoader{
		middleware:       middleware,
		validator:        validator,
		sanitizer:        sanitizer,
		cache:            NewConfigCache(30 * time.Minute),
		retryPolicy:      getDefaultRetryPolicy(),
		enableMonitoring: false,
	}
}

// LoadConfig loads configuration with comprehensive validation and caching
func (ecl *EnhancedConfigLoader) LoadConfig(ctx context.Context, options *ConfigLoadOptions) (*domain.Config, error) {
	if options == nil {
		options = getDefaultLoadOptions()
	}

	start := time.Now()

	// Check cache first
	if options.EnableCache && !options.ForceRefresh {
		if cached := ecl.cache.Get(); cached != nil {
			return cached, nil
		}
	}

	// Load configuration with retry
	config, err := ecl.loadConfigWithRetry(ctx, options)
	if err != nil {
		return nil, err
	}

	// Apply validation based on level
	validationResult := ecl.applyValidation(config, options.ValidationLevel)
	if !validationResult.IsValid {
		return nil, pkgerrors.HandleValidationError("LoadConfig",
			fmt.Errorf("validation failed: %s", ecl.formatValidationErrors(validationResult.Errors)))
	}

	// Apply sanitization if enabled
	if options.EnableSanitization {
		ecl.sanitizer.SanitizeConfig(config, validationResult)
	}

	// Update cache
	if options.EnableCache {
		ecl.cache.Set(config)
	}

	duration := time.Since(start)
	if ecl.enableMonitoring {
		fmt.Printf("📊 Config loaded in %v (validation: %s, sanitization: %v)\n",
			duration, options.ValidationLevel, options.EnableSanitization)
	}

	return config, nil
}

// SaveConfig saves configuration with validation and cache update
func (ecl *EnhancedConfigLoader) SaveConfig(ctx context.Context, config *domain.Config, options *ConfigSaveOptions) (*domain.Config, error) {
	if options == nil {
		options = getDefaultSaveOptions()
	}

	// Validate before saving
	validationResult := ecl.applyValidation(config, ValidationLevelComprehensive)
	if !validationResult.IsValid {
		return nil, pkgerrors.HandleValidationError("SaveConfig",
			fmt.Errorf("validation failed: %s", ecl.formatValidationErrors(validationResult.Errors)))
	}

	// Apply sanitization if enabled
	if options.EnableSanitization {
		ecl.sanitizer.SanitizeConfig(config, validationResult)
	}

	// Save with retry
	err := ecl.saveConfigWithRetry(ctx, config, options)
	if err != nil {
		return nil, err
	}

	// Update cache
	ecl.cache.Set(config)

	if ecl.enableMonitoring {
		fmt.Printf("💾 Config saved successfully\n")
	}

	return config, nil
}

// ValidateConfig validates configuration at specified level
func (ecl *EnhancedConfigLoader) ValidateConfig(ctx context.Context, config *domain.Config, level ValidationLevel) *ValidationResult {
	return ecl.applyValidation(config, level)
}

// loadConfigWithRetry loads configuration with retry logic
func (ecl *EnhancedConfigLoader) loadConfigWithRetry(ctx context.Context, options *ConfigLoadOptions) (*domain.Config, error) {
	var lastErr error

	for attempt := 0; attempt <= ecl.retryPolicy.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := ecl.calculateDelay(attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		// Use existing Load function with timeout context
		timeoutCtx, cancel := context.WithTimeout(ctx, options.Timeout)
		config, err := LoadWithContext(timeoutCtx)
		cancel()

		if err == nil {
			return config, nil
		}

		lastErr = err

		// Don't retry on certain errors
		if !ecl.shouldRetry(err) {
			break
		}
	}

	return nil, pkgerrors.HandleConfigError("LoadConfig", lastErr)
}

// saveConfigWithRetry saves configuration with retry logic
func (ecl *EnhancedConfigLoader) saveConfigWithRetry(ctx context.Context, config *domain.Config, options *ConfigSaveOptions) error {
	var lastErr error

	for attempt := 0; attempt <= ecl.retryPolicy.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := ecl.calculateDelay(attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := Save(config)
		if err == nil {
			return nil
		}

		lastErr = err

		if !ecl.shouldRetry(err) {
			break
		}
	}

	return pkgerrors.HandleConfigError("SaveConfig", lastErr)
}

// shouldRetry determines if an error should trigger a retry
func (ecl *EnhancedConfigLoader) shouldRetry(err error) bool {
	// Don't retry on certain errors
	if err.Error() == "validation failed" {
		return false
	}

	// Retry on temporary errors
	return true
}

// calculateDelay calculates exponential backoff delay
func (ecl *EnhancedConfigLoader) calculateDelay(attempt int) time.Duration {
	delay := ecl.retryPolicy.InitialDelay
	for i := 1; i < attempt; i++ {
		calculatedDelay := time.Duration(float64(delay) * ecl.retryPolicy.BackoffFactor)
		if calculatedDelay < ecl.retryPolicy.MaxDelay {
			delay = calculatedDelay
		} else {
			delay = ecl.retryPolicy.MaxDelay
		}
	}
	return delay
}

// formatValidationErrors formats validation errors into a readable string
func (ecl *EnhancedConfigLoader) formatValidationErrors(errors []ValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	message := fmt.Sprintf("Validation failed (%d errors):", len(errors))
	for i, err := range errors {
		message += fmt.Sprintf("\n%d. %s: %s", i+1, err.Field, err.Message)
	}
	return message
}

// Default option constructors

func getDefaultLoadOptions() *ConfigLoadOptions {
	return &ConfigLoadOptions{
		ForceRefresh:       false,
		EnableCache:        true,
		EnableSanitization: true,
		ValidationLevel:    ValidationLevelComprehensive,
		Timeout:            30 * time.Second,
	}
}

func getDefaultSaveOptions() *ConfigSaveOptions {
	return &ConfigSaveOptions{
		EnableSanitization: true,
		BackupEnabled:      true,
		ValidationLevel:    ValidationLevelComprehensive,
		CreateBackup:       false,
	}
}

func getDefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:    3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
	}
}
