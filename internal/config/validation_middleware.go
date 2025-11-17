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
	// Ensure sanitized data has embedded domain config
	if sanitizationResult.Sanitized.Config == nil {
		sanitizationResult.Sanitized.Config = sanitizationResult.Original.Config
	}

	return &ValidationResult{
		IsValid:   validationResult.IsValid,
		Errors:    append(validationResult.Errors, vm.convertSanitizationErrors(sanitizationResult)...),
		Warnings:  append(validationResult.Warnings, vm.convertSanitizationWarnings(sanitizationResult)...),
		Sanitized: sanitizationResult.Sanitized,
		Duration:  duration,
		Timestamp: time.Now(),
	}
}

// convertSanitizationErrors converts real sanitization errors to validation errors
func (vm *ValidationMiddleware) convertSanitizationErrors(result *SanitizationResult) []ValidationError {
	errors := []ValidationError{}
	for _, change := range result.Changes {
		if change.OldValue != change.NewValue {
			errors = append(errors, ValidationError{
				Field:    change.Field,
				Rule:     "sanitization",
				Value:    change.OldValue,
				Message:  fmt.Sprintf("Field %s was modified during sanitization", change.Field),
				Severity: SeverityWarning,
				Context: &ValidationContext{
					Metadata: map[string]string{
						"old_value": fmt.Sprintf("%v", change.OldValue),
						"new_value": fmt.Sprintf("%v", change.NewValue),
						"reason":    change.Reason,
					},
				},
				Timestamp: time.Now(),
			})
		}
	}
	return errors
}

// convertSanitizationWarnings converts real sanitization warnings to validation warnings
func (vm *ValidationMiddleware) convertSanitizationWarnings(result *SanitizationResult) []ValidationWarning {
	warnings := []ValidationWarning{}
	for _, warning := range result.Warnings {
		warnings = append(warnings, ValidationWarning{
			Field:   warning.Field,
			Message: warning.Message,
			Context: &ValidationContext{
				Metadata: map[string]string{
					"warning_type": fmt.Sprintf("%v", warning.Original),
					"suggestion":   warning.Reason,
				},
			},
			Timestamp: time.Now(),
		})
	}
	return warnings
}

// loadConfigWithValidation loads configuration from file with path validation
func (vm *ValidationMiddleware) loadConfigWithValidation(ctx context.Context, path string) (*domain.Config, error) {
	if path == "" {
		return nil, errors.NewError(errors.ErrConfigValidation, "Configuration path cannot be empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.NewError(errors.ErrConfigLoad, fmt.Sprintf("Failed to read configuration file: %s", err))
	}

	var config domain.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.NewError(errors.ErrConfigValidation, fmt.Sprintf("Failed to parse configuration file: %s", err))
	}

	return &config, nil
}

// saveConfig saves configuration to file
func (vm *ValidationMiddleware) saveConfig(ctx context.Context, config *domain.Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
