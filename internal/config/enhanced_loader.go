package config

import (
	"context"
	"fmt"
	"slices"
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

// ConfigCache provides configuration caching with TTL
type ConfigCache struct {
	config    *domain.Config
	loadedAt  time.Time
	ttl       time.Duration
	validator *ConfigValidator
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

// GetConfigSchema returns the configuration schema for validation
func (ecl *EnhancedConfigLoader) GetConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Version:     "1.0.0",
		Title:       "Clean Wizard Configuration Schema",
		Description: "Comprehensive configuration schema for clean-wizard",
		Types:       ecl.generateSchemaTypes(),
		Validation:  ecl.validator.rules,
	}
}

// ConfigSchema represents the configuration schema
type ConfigSchema struct {
	Version     string                 `json:"version"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Types       map[string]SchemaType  `json:"types"`
	Validation  *ConfigValidationRules `json:"validation"`
}

// SchemaType represents a type definition in the schema
type SchemaType struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Required    bool                   `json:"required"`
	Properties  map[string]*SchemaType `json:"properties,omitempty"`
	Items       *SchemaType            `json:"items,omitempty"`
	Enum        []any                  `json:"enum,omitempty"`
	Pattern     string                 `json:"pattern,omitempty"`
	Minimum     *float64               `json:"minimum,omitempty"`
	Maximum     *float64               `json:"maximum,omitempty"`
}

// ConfigSaveOptions provides options for configuration saving
type ConfigSaveOptions struct {
	EnableSanitization bool            `json:"enable_sanitization"`
	BackupEnabled      bool            `json:"backup_enabled"`
	ValidationLevel    ValidationLevel `json:"validation_level"`
	CreateBackup       bool            `json:"create_backup"`
}

// Internal methods

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

func (ecl *EnhancedConfigLoader) applyValidation(config *domain.Config, level ValidationLevel) *ValidationResult {
	switch level {
	case ValidationLevelNone:
		return &ValidationResult{IsValid: true, Timestamp: time.Now()}
	case ValidationLevelBasic:
		return ecl.validator.ValidateConfig(config) // Use existing validator
	case ValidationLevelComprehensive:
		// Add additional validation rules
		result := ecl.validator.ValidateConfig(config)
		ecl.applyComprehensiveValidation(config, result)
		return result
	case ValidationLevelStrict:
		// Apply all validation including strict checks
		result := ecl.validator.ValidateConfig(config)
		ecl.applyComprehensiveValidation(config, result)
		ecl.applyStrictValidation(config, result)
		return result
	default:
		return ecl.validator.ValidateConfig(config)
	}
}

func (ecl *EnhancedConfigLoader) applyComprehensiveValidation(config *domain.Config, result *ValidationResult) {
	// Additional comprehensive validation rules

	// Check for configuration consistency
	if config.SafeMode && ecl.hasCriticalRiskOperations(config) {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      "safe_mode",
			Message:    "Safe mode is enabled but critical risk operations exist",
			Suggestion: "Review critical operations or consider increasing risk tolerance",
		})
	}

	// Check for performance implications
	if len(config.Profiles) > 20 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      "profiles",
			Message:    "Large number of profiles may impact performance",
			Suggestion: "Consider consolidating similar profiles",
		})
	}
}

func (ecl *EnhancedConfigLoader) applyStrictValidation(config *domain.Config, result *ValidationResult) {
	// Strict validation rules that might fail

	// Require explicit profiles (no auto-generation)
	if len(config.Profiles) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "profiles",
			Rule:     "strict",
			Value:    config.Profiles,
			Message:  "Strict mode requires at least one explicit profile",
			Severity: SeverityError,
		})
		result.IsValid = false
	}

	// Require specific protected paths
	requiredPaths := []string{"/System", "/Library"}
	for _, required := range requiredPaths {
		if !ecl.isPathProtected(config.Protected, required) {
			result.Errors = append(result.Errors, ValidationError{
				Field:    "protected",
				Rule:     "strict",
				Value:    config.Protected,
				Message:  fmt.Sprintf("Strict mode requires path: %s", required),
				Severity: SeverityError,
			})
			result.IsValid = false
		}
	}
}

func (ecl *EnhancedConfigLoader) hasCriticalRiskOperations(config *domain.Config) bool {
	for _, profile := range config.Profiles {
		for _, op := range profile.Operations {
			if op.RiskLevel == domain.RiskCritical {
				return true
			}
		}
	}
	return false
}

func (ecl *EnhancedConfigLoader) isPathProtected(protected []string, target string) bool {
	return slices.Contains(protected, target)
}

func (ecl *EnhancedConfigLoader) shouldRetry(err error) bool {
	// Don't retry on certain errors
	if err.Error() == "validation failed" {
		return false
	}

	// Retry on temporary errors
	return true
}

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

func (ecl *EnhancedConfigLoader) generateSchemaTypes() map[string]SchemaType {
	return map[string]SchemaType{
		"Config": {
			Type:        "object",
			Description: "Main configuration structure",
			Required:    true,
			Properties: map[string]*SchemaType{
				"version": {
					Type:        "string",
					Description: "Configuration version",
					Required:    true,
					Pattern:     "^\\d+\\.\\d+\\.\\d+$",
				},
				"safe_mode": {
					Type:        "boolean",
					Description: "Enable safe mode",
					Required:    true,
				},
				"max_disk_usage": {
					Type:        "integer",
					Description: "Maximum disk usage percentage",
					Required:    true,
					Minimum:     func() *float64 { v := 10.0; return &v }(),
					Maximum:     func() *float64 { v := 95.0; return &v }(),
				},
				"protected": {
					Type:        "array",
					Description: "Protected paths",
					Required:    true,
					Items: &SchemaType{
						Type:    "string",
						Pattern: "^/.*",
					},
				},
				"profiles": {
					Type:        "object",
					Description: "Cleaning profiles",
					Required:    true,
				},
			},
		},
	}
}

// Helper constructors

func NewConfigCache(ttl time.Duration) *ConfigCache {
	return &ConfigCache{
		ttl:       ttl,
		validator: NewConfigValidator(),
	}
}

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

func (cc *ConfigCache) Set(config *domain.Config) {
	cc.config = config
	cc.loadedAt = time.Now()
}

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
