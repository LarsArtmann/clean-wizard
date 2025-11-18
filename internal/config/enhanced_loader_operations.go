package config

import (
	"context"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	pkgerrors "github.com/LarsArtmann/clean-wizard/internal/pkg/errors"
)

// ConfigLoadOptions provides options for configuration loading
type ConfigLoadOptions struct {
	ForceRefresh       bool                    `json:"force_refresh"`
	EnableCache        bool                    `json:"enable_cache"`
	EnableSanitization bool                    `json:"enable_sanitization"`
	ValidationLevel    domain.ValidationLevelType `json:"validation_level"`
	Timeout            time.Duration           `json:"timeout"`
}

// ConfigSaveOptions provides options for configuration saving
type ConfigSaveOptions struct {
	EnableSanitization bool                    `json:"enable_sanitization"`
	BackupEnabled      bool                    `json:"backup_enabled"`
	ValidationLevel    domain.ValidationLevelType `json:"validation_level"`
	CreateBackup       bool                    `json:"create_backup"`
}

// RetryPolicy defines retry behavior for configuration operations
type RetryPolicy struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
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
		return nil, pkgerrors.HandleConfigError("LoadConfig", err)
	}

	// Apply validation based on level
	validationResult := ecl.applyValidation(config, options.ValidationLevel)
	if !validationResult.IsValid {
		return nil, pkgerrors.HandleValidationError("LoadConfig",
			pkgerrors.NewValidationError("validation failed", ecl.formatValidationErrors(validationResult.Errors)))
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
	validationResult := ecl.applyValidation(config, options.ValidationLevel)
	if !validationResult.IsValid {
		return nil, pkgerrors.HandleValidationError("SaveConfig",
			pkgerrors.NewValidationError("validation failed", ecl.formatValidationErrors(validationResult.Errors)))
	}

	// Create backup if requested
	if (options.CreateBackup || options.BackupEnabled) && ecl.enableMonitoring {
		if err := ecl.createBackup(ctx, config); err != nil {
			fmt.Printf("⚠️ Backup failed: %v\n", err)
		}
	}

	// Apply sanitization if enabled
	if options.EnableSanitization {
		ecl.sanitizer.SanitizeConfig(config, validationResult)
	}

	// Save with retry
	err := ecl.saveConfigWithRetry(ctx, config, options)
	if err != nil {
		return nil, pkgerrors.HandleConfigError("SaveConfig", err)
	}

	// Update cache
	ecl.cache.Set(config)

	if ecl.enableMonitoring {
		fmt.Printf("💾 Config saved successfully\n")
	}

	return config, nil
}

// ValidateConfig validates configuration at specified level
func (ecl *EnhancedConfigLoader) ValidateConfig(ctx context.Context, config *domain.Config, level domain.ValidationLevelType) *ValidationResult {
	return ecl.applyValidation(config, level)
}