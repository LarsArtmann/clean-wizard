package config

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// applyValidation applies validation at the specified level
func (ecl *EnhancedConfigLoader) applyValidation(ctx context.Context, config *config.Config, level ValidationLevel) *ValidationResult {
	switch level {
	case ValidationLevelNone:
		return &ValidationResult{IsValid: true, Timestamp: time.Now()}
	case ValidationLevelBasic:
		return ecl.validator.ValidateConfig(config) // Use existing validator
	case ValidationLevelComprehensive:
		// Add additional validation rules
		result := ecl.validator.ValidateConfig(config)
		ecl.applyComprehensiveValidation(config, result)
		return result
	case ValidationLevelStrict:
		// Apply all validation including strict checks
		result := ecl.validator.ValidateConfig(config)
		ecl.applyComprehensiveValidation(config, result)
		ecl.applyStrictValidation(config, result)
		return result
	default:
		return ecl.validator.ValidateConfig(config)
	}
}

// applyComprehensiveValidation applies comprehensive validation rules
func (ecl *EnhancedConfigLoader) applyComprehensiveValidation(config *config.Config, result *ValidationResult) {
	// Additional comprehensive validation rules

	// Check for configuration consistency
	if ConfigSafetyLevel >= shared.SafetyLevelEnabled && ecl.hasCriticalRiskOperations(config) {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      "safety_level",
			Message:    "Safety level is enabled but critical risk operations exist",
			Suggestion: "Review critical operations or consider increasing risk tolerance",
		})
	}

	// Check for performance implications
	if len(ConfigProfiles) > 20 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      "profiles",
			Message:    "Large number of profiles may impact performance",
			Suggestion: "Consider consolidating similar profiles",
		})
	}
}

// applyStrictValidation applies strict validation rules
func (ecl *EnhancedConfigLoader) applyStrictValidation(config *config.Config, result *ValidationResult) {
	// Strict validation rules that might fail

	// Require explicit profiles (no auto-generation)
	if len(ConfigProfiles) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "profiles",
			Rule:     "strict",
			Value:    ConfigProfiles,
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
			requiredPaths = []string{"/System", "/Library"} // Final fallback
		}
	}
	for _, required := range requiredPaths {
		if !ecl.isPathProtected(ConfigProtected, required) {
			result.Errors = append(result.Errors, ValidationError{
				Field:    "protected",
				Rule:     "strict",
				Value:    ConfigProtected,
				Message:  fmt.Sprintf("Strict mode requires path: %s", required),
				Severity: SeverityError,
			})
			result.IsValid = false
		}
	}
}

// hasCriticalRiskOperations checks if config contains critical risk operations
func (ecl *EnhancedConfigLoader) hasCriticalRiskOperations(config *config.Config) bool {
	for _, profile := range ConfigProfiles {
		// Guard against nil profiles (e.g., from "profile: null" in YAML)
		if profile == nil {
			continue
		}
		for _, op := range profile.Operations {
			if op.RiskLevel == shared.RiskCritical {
				return true
			}
		}
	}
	return false
}

// isPathProtected checks if a path is in the protected list
func (ecl *EnhancedConfigLoader) isPathProtected(protected []string, target string) bool {
	return slices.Contains(protected, target)
}

// formatValidationErrors formats validation errors for display
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
