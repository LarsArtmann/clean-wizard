package config

import (
	"context"
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	pkgerrors "github.com/LarsArtmann/clean-wizard/internal/shared/utils/pkg/errors"
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
	validationResult := ecl.applyValidation(ctx, config, ValidationLevel(options.ValidationLevel))
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

	// Create backup if requested (independent of monitoring)
	if options.CreateBackup {
		if err := ecl.createBackup(ctx, config); err != nil {
			if ecl.enableMonitoring {
				fmt.Printf("⚠️ Backup failed: %v\n", err)
			}
		} else if ecl.enableMonitoring {
			fmt.Printf("💾 Configuration backup created\n")
		}
	}

	// Always run validation at requested level
	validationResult := ecl.applyValidation(ctx, config, ValidationLevel(options.ValidationLevel))

	// Apply sanitization if enabled (after initial validation)
	if options.EnableSanitization {
		ecl.sanitizer.SanitizeConfig(config, validationResult)
		// Re-check validity after sanitization
		validationResult = ecl.applyValidation(ctx, config, ValidationLevel(options.ValidationLevel))
	}

	// Check final validation state
	if !validationResult.IsValid {
		if options.ForceSave {
			if ecl.enableMonitoring {
				fmt.Printf("⚠️ Validation failed, forcing save: %s\n", ecl.formatValidationErrors(validationResult.Errors))
			}
		} else {
			return nil, fmt.Errorf("validation failed: %s", ecl.formatValidationErrors(validationResult.Errors))
		}
	}

	// Only call saveConfigWithRetry when validation permits saving
	err := ecl.saveConfigWithRetry(ctx, config, options)
	if err != nil {
		return nil, err
	}

	// Update cache
	ecl.cache.Set(config)

	if ecl.enableMonitoring {
		if options.ForceSave && !validationResult.IsValid {
			fmt.Printf("💾 Config saved (forced) despite validation failures\n")
		} else {
			fmt.Printf("💾 Config saved successfully\n")
		}
	}

	return config, nil
}

// ValidateConfig validates configuration at specified level
func (ecl *EnhancedConfigLoader) ValidateConfig(ctx context.Context, config *domain.Config, level domain.ValidationLevelType) *ValidationResult {
	return ecl.applyValidation(ctx, config, ValidationLevel(level))
}
