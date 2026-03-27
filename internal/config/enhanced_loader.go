package config

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// applyValidation applies validation at the specified level.
func (ecl *EnhancedConfigLoader) applyValidation(
	ctx context.Context,
	config *domain.Config,
	level domain.ValidationLevelType,
) *ValidationResult {
	switch level {
	case domain.ValidationLevelNoneType:
		return &ValidationResult{IsValid: true, Timestamp: time.Now()}
	case domain.ValidationLevelBasicType:
		return ecl.validator.ValidateConfig(config) // Use existing validator
	case domain.ValidationLevelComprehensiveType:
		// Add additional validation rules
		result := ecl.validator.ValidateConfig(config)
		ecl.applyComprehensiveValidation(config, result)

		return result
	case domain.ValidationLevelStrictType:
		// Apply all validation including strict checks
		result := ecl.validator.ValidateConfig(config)
		ecl.applyComprehensiveValidation(config, result)
		ecl.applyStrictValidation(config, result)

		return result
	default:
		return ecl.validator.ValidateConfig(config)
	}
}

// applyComprehensiveValidation applies comprehensive validation rules.
func (ecl *EnhancedConfigLoader) applyComprehensiveValidation(
	config *domain.Config,
	result *ValidationResult,
) {
	// Additional comprehensive validation rules

	// Check for configuration consistency
	if config.SafeMode.IsEnabled() && ecl.hasCriticalRiskOperations(config) {
		result.Warnings = append(result.Warnings, *ecl.createSafeModeWarning())
	}

	// Check for performance implications
	if len(config.Profiles) > MaxProfileCountWarning {
		result.Warnings = append(result.Warnings, *ecl.createProfilesWarning())
	}
}

// applyStrictValidation applies strict validation rules.
func (ecl *EnhancedConfigLoader) applyStrictValidation(
	config *domain.Config,
	result *ValidationResult,
) {
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
	requiredPaths := ecl.validator.rules.ProtectedSystemPaths
	// Fallback to DefaultProtectedPaths if rules slice is nil/empty (maintaining consistency with sanitizer)
	if len(requiredPaths) == 0 {
		requiredPaths = ecl.validator.rules.DefaultProtectedPaths
		if len(requiredPaths) == 0 {
			requiredPaths = []string{domain.PathSystem, domain.PathLibrary} // Final fallback
		}
	}

	for _, required := range requiredPaths {
		if !ecl.isPathProtected(config.Protected, required) {
			result.Errors = append(result.Errors, ValidationError{
				Field:    "protected",
				Rule:     "strict",
				Value:    config.Protected,
				Message:  "Strict mode requires path: " + required,
				Severity: SeverityError,
			})
			result.IsValid = false
		}
	}
}

// hasCriticalRiskOperations checks if config contains critical risk operations.
func (ecl *EnhancedConfigLoader) hasCriticalRiskOperations(config *domain.Config) bool {
	for _, profile := range config.Profiles {
		// Guard against nil profiles (e.g., from "profile: null" in YAML)
		if profile == nil {
			continue
		}

		for _, op := range profile.Operations {
			if op.RiskLevel == domain.RiskLevelType(domain.RiskLevelCriticalType) {
				return true
			}
		}
	}

	return false
}

// isPathProtected checks if a path is in the protected list.
func (ecl *EnhancedConfigLoader) isPathProtected(protected []string, target string) bool {
	return slices.Contains(protected, target)
}

// formatValidationErrors formats validation errors for display.
func (ecl *EnhancedConfigLoader) formatValidationErrors(errors []ValidationError) string {
	if len(errors) == 0 {
		return "no errors"
	}

	var messages []string
	for _, err := range errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}

	return fmt.Sprintf("%d validation errors: %s", len(errors), fmt.Sprintf("%v", messages))
}

// createSafeModeWarning creates a warning for safe mode with critical risk operations.
func (ecl *EnhancedConfigLoader) createSafeModeWarning() *ValidationWarning {
	return &ValidationWarning{
		Field:      "safe_mode",
		Message:    "Safe mode is enabled but configuration contains critical risk operations",
		Suggestion: "Consider disabling safe mode or reviewing critical operations",
	}
}

// createProfilesWarning creates a warning for excessive profile count.
func (ecl *EnhancedConfigLoader) createProfilesWarning() *ValidationWarning {
	return &ValidationWarning{
		Field:      "profiles",
		Message:    "Large number of profiles may impact performance",
		Suggestion: "Consider consolidating profiles or reducing count below 20",
	}
}
