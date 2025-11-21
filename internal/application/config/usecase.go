package config

import (
	"context"
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/domain/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
	pkgerrors "github.com/LarsArtmann/clean-wizard/internal/shared/utils/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// ApplyConfigurationFixes applies all configuration fixing business logic
// This is the main use case for configuration restoration and repair
func ApplyConfigurationFixes(config *config.Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Create viper instance for fixer
	v := viper.New()
	
	// Apply comprehensive configuration fixes
	fixer := NewConfigFixer(v)
	fixer.FixAll(config)

	// Apply comprehensive validation with strict enforcement
	if validator := NewConfigValidator(); validator != nil {
		validationResult := validator.ValidateConfig(config)
		if !validationResult.IsValid {
			// CRITICAL: Fail fast on validation errors for production safety
			for _, err := range validationResult.Errors {
				log.Error().
					Str("field", err.Field).
					Err(fmt.Errorf("%s", err.Message)).
					Msg("Configuration validation error")
			}
			return fmt.Errorf("configuration validation failed with %d errors", len(validationResult.Errors))
		}
	}

	return nil
}

// ValidateConfigurationOnly validates configuration without applying fixes
// Use case for read-only validation scenarios
func ValidateConfigurationOnly(config *config.Config) (*config.ConfigValidationResult, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if validator := NewConfigValidator(); validator != nil {
		validationResult := validator.ValidateConfig(config)
		return validationResult, nil
	}

	return &ConfigValidationResult{
		IsValid: true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
	}, nil
}
