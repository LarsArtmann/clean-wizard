package config

import (
	"fmt"
	"slices"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// applyValidation applies validation at the specified level
func (ecl *EnhancedConfigLoader) applyValidation(config *domain.Config, level ValidationLevel) *ValidationResult {
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
func (ecl *EnhancedConfigLoader) applyComprehensiveValidation(config *domain.Config, result *ValidationResult) {
	// Additional comprehensive validation rules

	// Check for configuration consistency
	if config.SafeMode && ecl.hasCriticalRiskOperations(config) {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      "safe_mode",
			Message:    "Safe mode is enabled but critical risk operations exist",
			Suggestion: "Review critical operations or consider increasing risk tolerance",
		})
	}

	// Check for performance implications
	if len(config.Profiles) > 20 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      "profiles",
			Message:    "Large number of profiles may impact performance",
			Suggestion: "Consider consolidating similar profiles",
		})
	}
}

// applyStrictValidation applies strict validation rules
func (ecl *EnhancedConfigLoader) applyStrictValidation(config *domain.Config, result *ValidationResult) {
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
	// Fallback to literals if rules slice is nil/empty
	if len(requiredPaths) == 0 {
		requiredPaths = []string{"/System", "/Library"}
	}
	for _, required := range requiredPaths {
		if !ecl.isPathProtected(config.Protected, required) {
			result.Errors = append(result.Errors, ValidationError{
				Field:    "protected",
				Rule:     "strict",
				Value:    config.Protected,
				Message:  fmt.Sprintf("Strict mode requires path: %s", required),
				Severity: SeverityError,
			})
			result.IsValid = false
		}
	}
}

// hasCriticalRiskOperations checks if config contains critical risk operations
func (ecl *EnhancedConfigLoader) hasCriticalRiskOperations(config *domain.Config) bool {
	for _, profile := range config.Profiles {
		// Guard against nil profiles (e.g., from "profile: null" in YAML)
		if profile == nil {
			continue
		}
		for _, op := range profile.Operations {
			if op.RiskLevel == domain.RiskCritical {
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

// mapValidatorRulesToSchemaRules converts validator rules to stable schema rules
func (ecl *EnhancedConfigLoader) mapValidatorRulesToSchemaRules() *ConfigValidationRules {
	rules := ecl.validator.rules
	
	// Create a copy to prevent external modifications to internal state
	schemaRules := &ConfigValidationRules{
		// Numeric Constraints
		MaxDiskUsage:      rules.MaxDiskUsage,
		MinProtectedPaths: rules.MinProtectedPaths,
		MaxProfiles:       rules.MaxProfiles,
		MaxOperations:     rules.MaxOperations,
		
		// String Constraints
		ProfileNamePattern: rules.ProfileNamePattern,
		PathPattern:        rules.PathPattern,
		
		// Array Constraints
		UniquePaths:    rules.UniquePaths,
		UniqueProfiles: rules.UniqueProfiles,
		
		// Safety Constraints
		ProtectedSystemPaths: make([]string, len(rules.ProtectedSystemPaths)),
		RequireSafeMode:      rules.RequireSafeMode,
		
		// Risk Constraints
		MaxRiskLevel:   rules.MaxRiskLevel,
		BackupRequired:  rules.BackupRequired,
	}
	
	// Deep copy slice to prevent modifications
	copy(schemaRules.ProtectedSystemPaths, rules.ProtectedSystemPaths)
	
	return schemaRules
}

// getSchemaMinimum returns the minimum value for max_disk_usage from rules
func (ecl *EnhancedConfigLoader) getSchemaMinimum() *float64 {
	if ecl.validator.rules.MaxDiskUsage != nil && ecl.validator.rules.MaxDiskUsage.Min != nil {
		v := float64(*ecl.validator.rules.MaxDiskUsage.Min)
		return &v
	}
	v := 10.0 // fallback to current literal
	return &v
}

// getSchemaMaximum returns the maximum value for max_disk_usage from rules
func (ecl *EnhancedConfigLoader) getSchemaMaximum() *float64 {
	if ecl.validator.rules.MaxDiskUsage != nil && ecl.validator.rules.MaxDiskUsage.Max != nil {
		v := float64(*ecl.validator.rules.MaxDiskUsage.Max)
		return &v
	}
	v := 95.0 // fallback to current literal
	return &v
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