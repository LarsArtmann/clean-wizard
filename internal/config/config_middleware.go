package config

import (
	"context"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/pkg/errors"
)

// ConfigMiddleware provides core configuration validation and loading
type ConfigMiddleware struct {
	validator *ConfigValidator
	sanitizer *ConfigSanitizer
	logger    ValidationLogger
}

// NewConfigMiddleware creates a new validation middleware
func NewConfigMiddleware() *ConfigMiddleware {
	return &ConfigMiddleware{
		validator: NewConfigValidator(),
		sanitizer: NewConfigSanitizer(),
		logger:    NewDefaultValidationLogger(false),
	}
}

// NewConfigMiddlewareWithLogger creates middleware with custom logger
func NewConfigMiddlewareWithLogger(logger ValidationLogger) *ConfigMiddleware {
	return &ConfigMiddleware{
		validator: NewConfigValidator(),
		sanitizer: NewConfigSanitizer(),
		logger:    logger,
	}
}

// ValidateAndLoadConfig validates and loads configuration
func (cm *ConfigMiddleware) ValidateAndLoadConfig(ctx context.Context) (*domain.Config, error) {
	start := time.Now()
	
	// Load configuration with validation
	config, err := cm.loadConfigWithValidation(ctx)
	if err != nil {
		cm.logger.LogError("config", "load", err)
		return nil, err
	}

	// Log successful validation
	validationResult := cm.validator.ValidateConfig(config)
	cm.logger.LogValidation(validationResult)

	if !validationResult.IsValid {
		return nil, errors.ValidationError("configuration validation failed", validationResult.Errors)
	}

	// Apply sanitization if configured
	sanitizationResult := cm.sanitizer.SanitizeConfig(config)
	cm.logger.LogSanitization(sanitizationResult)

	duration := time.Since(start)
	if cm.logger.(*DefaultValidationLogger).enableDetailedLogging {
		cm.logger.LogValidation(&ValidationResult{
			IsValid:   true,
			Errors:    []ValidationError{},
			Warnings:  []ValidationWarning{},
			Sanitized: map[string]any{
				"sanitized": sanitizationResult.Sanitized,
				"changes":   sanitizationResult.Changes,
			},
			Duration:  duration,
			Timestamp: time.Now(),
		})
	}

	return config, nil
}

// ValidateAndSaveConfig validates and saves configuration
func (cm *ConfigMiddleware) ValidateAndSaveConfig(ctx context.Context, cfg *domain.Config) (*domain.Config, error) {
	start := time.Now()

	// Validate configuration
	validationResult := cm.validator.ValidateConfig(cfg)
	cm.logger.LogValidation(validationResult)

	if !validationResult.IsValid {
		cm.logger.LogError("config", "save", 
			errors.ValidationError("configuration validation failed", validationResult.Errors))
		return nil, errors.ValidationError("configuration validation failed", validationResult.Errors)
	}

	// Apply sanitization
	sanitizationResult := cm.sanitizer.SanitizeConfig(cfg)
	cm.logger.LogSanitization(sanitizationResult)

	// Save configuration
	if err := cm.saveConfig(ctx, cfg); err != nil {
		cm.logger.LogError("config", "save", err)
		return nil, err
	}

	duration := time.Since(start)
	if cm.logger.(*DefaultValidationLogger).enableDetailedLogging {
		cm.logger.LogValidation(&ValidationResult{
			IsValid:   true,
			Errors:    []ValidationError{},
			Warnings:  []ValidationWarning{},
			Sanitized: map[string]any{
				"sanitized": sanitizationResult.Sanitized,
				"changes":   sanitizationResult.Changes,
			},
			Duration:  duration,
			Timestamp: time.Now(),
		})
	}

	return cfg, nil
}

// loadConfigWithValidation loads configuration with validation
func (cm *ConfigMiddleware) loadConfigWithValidation(ctx context.Context) (*domain.Config, error) {
	// This would load from file, database, etc.
	// For now, return a basic configuration
	return &domain.Config{
		Version:    "1.0.0",
		SafeMode:   true,
		MaxDiskUsage: 50,
		Protected:  []string{"/", "/System", "/Library", "/usr", "/etc"},
		Profiles:   map[string]*domain.Profile{},
	}, nil
}

// saveConfig saves configuration
func (cm *ConfigMiddleware) saveConfig(ctx context.Context, cfg *domain.Config) error {
	// This would save to file, database, etc.
	// For now, just log
	return nil
}