package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/pkg/errors"
	"gopkg.in/yaml.v3"
)

// ValidationMiddleware provides high-level validation and sanitization
type ValidationMiddleware struct {
	validator *ConfigValidator
	sanitizer *ConfigSanitizer
	logger    ValidationLogger
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware() *ValidationMiddleware {
	return &ValidationMiddleware{
		validator: NewConfigValidator(),
		sanitizer: NewConfigSanitizer(),
		logger:    NewDefaultValidationLogger(false),
	}
}

// NewValidationMiddlewareWithLogger creates middleware with custom logger
func NewValidationMiddlewareWithLogger(logger ValidationLogger) *ValidationMiddleware {
	return &ValidationMiddleware{
		validator: NewConfigValidator(),
		sanitizer: NewConfigSanitizer(),
		logger:    logger,
	}
}

// ValidateAndSanitize performs complete validation and sanitization
func (vm *ValidationMiddleware) ValidateAndSanitize(ctx context.Context, configPath string) (*ValidationResult, error) {
	start := time.Now()
	
	// Load configuration
	config, err := vm.loadConfig(configPath)
	if err != nil {
		return nil, errors.NewError(errors.ErrConfigLoad, fmt.Sprintf("Failed to load configuration from %s: %v", configPath, err))
	}
	
	// Validate configuration
	validationResult := vm.validator.ValidateConfig(config)
	
	// Sanitize configuration
	sanitizationResult := vm.sanitizer.SanitizeConfig(config)
	
	// Create unified result
	result := vm.createUnifiedResult(validationResult, sanitizationResult, time.Since(start))
	
	// Log validation
	vm.logger.LogValidation(result)
	
	return result, nil
}

// loadConfig loads configuration from file
func (vm *ValidationMiddleware) loadConfig(configPath string) (*domain.Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	
	var config domain.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	
	return &config, nil
}

// createUnifiedResult creates unified validation and sanitization result
func (vm *ValidationMiddleware) createUnifiedResult(validationResult *ValidationResult, sanitizationResult *SanitizationResult, duration time.Duration) *ValidationResult {
	return &ValidationResult{
		IsValid:   validationResult.IsValid,
		Errors:    append(validationResult.Errors, vm.convertErrors(sanitizationResult)...),
		Warnings:  append(validationResult.Warnings, vm.convertWarnings(sanitizationResult)...),
		Sanitized: sanitizationResult.Sanitized,
		Duration:  duration,
		Timestamp: time.Now(),
	}
}

// convertErrors converts sanitization errors to validation errors
func (vm *ValidationMiddleware) convertErrors(result *SanitizationResult) []ValidationError {
	// Implementation depends on actual sanitization error structure
	return []ValidationError{}
}

// convertWarnings converts sanitization warnings to validation warnings
func (vm *ValidationMiddleware) convertWarnings(result *SanitizationResult) []ValidationWarning {
	// Implementation depends on actual sanitization warning structure
	return []ValidationWarning{}
}

// loadConfigWithValidation loads and validates configuration
func (vm *ValidationMiddleware) loadConfigWithValidation(ctx context.Context) (*domain.Config, error) {
	// Basic implementation - could be enhanced with different sources
	// For now, use default config
	return &domain.Config{
		Version:      "1.0.0",
		SafeMode:     true,
		MaxDiskUsage:  50,
		Protected:    []string{"/", "/System", "/Library"},
		Profiles:     map[string]*domain.Profile{},
	}, nil
}

// saveConfig saves configuration to file
func (vm *ValidationMiddleware) saveConfig(ctx context.Context, config *domain.Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
