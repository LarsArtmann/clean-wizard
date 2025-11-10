package config

import (
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// OperationValidator provides validation for operation settings
type OperationValidator struct {
	enableDetailedValidation bool
}

// NewOperationValidator creates a new operation validator
func NewOperationValidator(enableDetailed bool) *OperationValidator {
	return &OperationValidator{
		enableDetailedValidation: enableDetailed,
	}
}

// ValidateOperationSettings validates operation-specific settings
func (ov *OperationValidator) ValidateOperationSettings(op *domain.CleanupOperation) error {
	if op.Settings == nil {
		return fmt.Errorf("operation '%s' has no settings", op.Name)
	}

	if !op.Settings.IsValid() {
		return fmt.Errorf("operation '%s' has invalid settings", op.Name)
	}

	// Type-specific validation
	switch op.Settings.Type {
	case domain.OperationTypeNixStore:
		return ov.validateNixStoreSettings(op.Settings.NixStore, op.RiskLevel)
	case domain.OperationTypeHomebrew:
		return ov.validateHomebrewSettings(op.Settings.Homebrew, op.RiskLevel)
	case domain.OperationTypePackageCache:
		return ov.validatePackageCacheSettings(op.Settings.Package, op.RiskLevel)
	case domain.OperationTypeTempFiles:
		return ov.validateTempFilesSettings(op.Settings.TempFiles, op.RiskLevel)
	default:
		return fmt.Errorf("unknown operation type: %s", op.Settings.Type)
	}
}

// validateNixStoreSettings validates Nix store settings
func (ov *OperationValidator) validateNixStoreSettings(settings *domain.NixStoreSettings, riskLevel domain.RiskLevel) error {
	if settings == nil {
		return fmt.Errorf("Nix store settings are required")
	}

	// Risk-based validation
	if riskLevel.IsHigherOrEqualThan(domain.RiskHigh) {
		if settings.KeepGenerations < 2 {
			return fmt.Errorf("high risk operations must keep at least 2 generations")
		}
		if settings.MinAge < 24*time.Hour {
			return fmt.Errorf("high risk operations must have minimum age of 24 hours")
		}
	}

	// Logic validation
	if settings.KeepGenerations < 1 || settings.KeepGenerations > 50 {
		return fmt.Errorf("keep_generations must be between 1 and 50, got: %d", settings.KeepGenerations)
	}

	if settings.MinAge < 0 {
		return fmt.Errorf("min_age cannot be negative, got: %v", settings.MinAge)
	}

	return nil
}

// validateHomebrewSettings validates Homebrew settings
func (ov *OperationValidator) validateHomebrewSettings(settings *domain.HomebrewSettings, riskLevel domain.RiskLevel) error {
	if settings == nil {
		return fmt.Errorf("Homebrew settings are required")
	}

	// Risk-based validation
	if riskLevel.IsHigherOrEqualThan(domain.RiskHigh) {
		if settings.KeepLatest < 1 {
			return fmt.Errorf("high risk operations must keep at least 1 latest version")
		}
		if settings.MinAge < 12*time.Hour {
			return fmt.Errorf("high risk operations must have minimum age of 12 hours")
		}
	}

	// Logic validation
	if settings.KeepLatest < 1 || settings.KeepLatest > 20 {
		return fmt.Errorf("keep_latest must be between 1 and 20, got: %d", settings.KeepLatest)
	}

	if settings.MinAge < 0 {
		return fmt.Errorf("min_age cannot be negative, got: %v", settings.MinAge)
	}

	return nil
}

// validatePackageCacheSettings validates package cache settings
func (ov *OperationValidator) validatePackageCacheSettings(settings *domain.PackageCacheSettings, riskLevel domain.RiskLevel) error {
	if settings == nil {
		return fmt.Errorf("Package cache settings are required")
	}

	// Risk-based validation
	if riskLevel.IsHigherOrEqualThan(domain.RiskHigh) {
		if settings.MaxSize > 0 && settings.MaxSize < 1024*1024*1024 { // 1GB
			return fmt.Errorf("high risk operations should not limit cache size too aggressively")
		}
	}

	// Logic validation
	if settings.MaxAge <= 0 {
		return fmt.Errorf("max_age must be positive, got: %v", settings.MaxAge)
	}

	if settings.MaxSize < 0 {
		return fmt.Errorf("max_size cannot be negative, got: %d", settings.MaxSize)
	}

	if len(settings.IncludeTypes) == 0 {
		return fmt.Errorf("include_types cannot be empty")
	}

	return nil
}

// validateTempFilesSettings validates temporary files settings
func (ov *OperationValidator) validateTempFilesSettings(settings *domain.TempFilesSettings, riskLevel domain.RiskLevel) error {
	if settings == nil {
		return fmt.Errorf("Temp files settings are required")
	}

	// Risk-based validation
	if riskLevel.IsHigherOrEqualThan(domain.RiskHigh) {
		if len(settings.Paths) == 0 {
			return fmt.Errorf("high risk operations must specify paths explicitly")
		}
		if settings.MaxAge < 1*time.Hour {
			return fmt.Errorf("high risk operations must have minimum age of 1 hour")
		}
	}

	// Logic validation
	if settings.MaxAge <= 0 {
		return fmt.Errorf("max_age must be positive, got: %v", settings.MaxAge)
	}

	if len(settings.Paths) == 0 && len(settings.Patterns) == 0 {
		return fmt.Errorf("must specify at least paths or patterns")
	}

	// Validate paths don't contain dangerous patterns
	for _, path := range settings.Paths {
		if len(path) > 1000 {
			return fmt.Errorf("path too long: %s", path)
		}
	}

	// Security validation
	for _, excludePath := range settings.ExcludePaths {
		if excludePath == "" {
			return fmt.Errorf("exclude paths cannot contain empty strings")
		}
	}

	return nil
}