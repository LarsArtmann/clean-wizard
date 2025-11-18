package config

import (
	"context"
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	pkgerrors "github.com/LarsArtmann/clean-wizard/internal/pkg/errors"
)

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
	validationResult := ecl.applyValidation(config, ValidationLevel(options.ValidationLevel))
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

	// Create backup if requested
	if (options.CreateBackup || options.BackupEnabled) && ecl.enableMonitoring {
		if err := ecl.createBackup(ctx, config); err != nil {
			fmt.Printf("⚠️ Backup failed: %v\n", err)
		}
	}

	// Apply sanitization if enabled
	if options.EnableSanitization {
		validationResult := ecl.applyValidation(config, ValidationLevel(options.ValidationLevel))
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
func (ecl *EnhancedConfigLoader) ValidateConfig(ctx context.Context, config *domain.Config, level domain.ValidationLevelType) *ValidationResult {
	return ecl.applyValidation(config, ValidationLevel(level))
}