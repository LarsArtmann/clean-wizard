package config

import (
	"context"
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	pkgerrors "github.com/LarsArtmann/clean-wizard/internal/pkg/errors"
)

// LoadConfig loads configuration with comprehensive validation and caching.
func (ecl *EnhancedConfigLoader) LoadConfig(ctx context.Context, options *ConfigLoadOptions) (*domain.Config, error) {
	if options == nil {
		options = getDefaultLoadOptions()
	}

	start := time.Now()

	// Check cache first
	if options.EnableCache == CacheOptionEnabled && options.ForceRefresh == RefreshOptionDisabled {
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
	validationResult := ecl.applyValidation(ctx, config, options.ValidationLevel)
	if !validationResult.IsValid {
		return nil, pkgerrors.HandleValidationError("LoadConfig",
			fmt.Errorf("validation failed: %s", ecl.formatValidationErrors(validationResult.Errors)))
	}

	// Apply sanitization if enabled
	if options.EnableSanitization == SanitizeOptionEnabled {
		ecl.sanitizer.SanitizeConfig(config, validationResult)
	}

	// Update cache
	if options.EnableCache == CacheOptionEnabled {
		ecl.cache.Set(config)
	}

	duration := time.Since(start)
	if ecl.enableMonitoring == MonitoringOptionEnabled {
		fmt.Printf("📊 Config loaded in %v (validation: %s, sanitization: %v)\n",
			duration, options.ValidationLevel, options.EnableSanitization)
	}

	return config, nil
}

// SaveConfig saves configuration with validation and cache update.
func (ecl *EnhancedConfigLoader) SaveConfig(ctx context.Context, config *domain.Config, options *ConfigSaveOptions) (*domain.Config, error) {
	if options == nil {
		options = getDefaultSaveOptions()
	}

	ecl.handleBackup(ctx, config, options)

	validationResult := ecl.validateAndSanitize(ctx, config, options)

	if err := ecl.handleValidationFailure(validationResult, options); err != nil {
		return nil, err
	}

	if err := ecl.saveConfigWithRetry(ctx, config, options); err != nil {
		return nil, err
	}

	ecl.cache.Set(config)
	ecl.logSaveResult(options, validationResult)

	return config, nil
}

// handleBackup creates a backup if requested.
func (ecl *EnhancedConfigLoader) handleBackup(ctx context.Context, config *domain.Config, options *ConfigSaveOptions) {
	if options.CreateBackup != BackupOptionEnabled && options.BackupEnabled != BackupOptionEnabled {
		return
	}

	err := ecl.createBackup(ctx, config)
	if ecl.enableMonitoring != MonitoringOptionEnabled {
		return
	}

	if err != nil {
		fmt.Printf("⚠️ Backup failed: %v\n", err)
	} else {
		fmt.Printf("💾 Configuration backup created\n")
	}
}

// validateAndSanitize validates and optionally sanitizes the config.
func (ecl *EnhancedConfigLoader) validateAndSanitize(ctx context.Context, config *domain.Config, options *ConfigSaveOptions) *ValidationResult {
	validationResult := ecl.applyValidation(ctx, config, options.ValidationLevel)

	if options.EnableSanitization == SanitizeOptionEnabled {
		ecl.sanitizer.SanitizeConfig(config, validationResult)
		validationResult = ecl.applyValidation(ctx, config, options.ValidationLevel)
	}

	return validationResult
}

// handleValidationFailure returns error if validation failed and save is not forced.
func (ecl *EnhancedConfigLoader) handleValidationFailure(validationResult *ValidationResult, options *ConfigSaveOptions) error {
	if validationResult.IsValid {
		return nil
	}

	if options.ForceSave == SaveOptionEnabled {
		if ecl.enableMonitoring == MonitoringOptionEnabled {
			fmt.Printf("⚠️ Validation failed, forcing save: %s\n", ecl.formatValidationErrors(validationResult.Errors))
		}
		return nil
	}

	return fmt.Errorf("validation failed: %s", ecl.formatValidationErrors(validationResult.Errors))
}

// logSaveResult logs the save result if monitoring is enabled.
func (ecl *EnhancedConfigLoader) logSaveResult(options *ConfigSaveOptions, validationResult *ValidationResult) {
	if ecl.enableMonitoring != MonitoringOptionEnabled {
		return
	}

	if options.ForceSave == SaveOptionEnabled && !validationResult.IsValid {
		fmt.Printf("💾 Config saved (forced) despite validation failures\n")
	} else {
		fmt.Printf("💾 Config saved successfully\n")
	}
}

// ValidateConfig validates configuration at specified level.
func (ecl *EnhancedConfigLoader) ValidateConfig(ctx context.Context, config *domain.Config, level domain.ValidationLevelType) *ValidationResult {
	return ecl.applyValidation(ctx, config, level)
}
