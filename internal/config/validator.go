package config

import (
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// ConfigValidator provides comprehensive type-safe configuration validation
type ConfigValidator struct {
	rules             *ConfigValidationRules
	basicValidator    *BasicValidator
	fieldValidator    *FieldValidator
	businessValidator *BusinessValidator
	securityValidator *SecurityValidator
}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator() *ConfigValidator {
	rules := GetDefaultValidationRules()
	return &ConfigValidator{
		rules:             rules,
		basicValidator:    NewBasicValidator(rules),
		fieldValidator:    NewFieldValidator(rules),
		businessValidator: NewBusinessValidator(rules),
		securityValidator: NewSecurityValidator(rules),
	}
}

// NewConfigValidatorWithRules creates a validator with custom rules
func NewConfigValidatorWithRules(rules *ConfigValidationRules) *ConfigValidator {
	return &ConfigValidator{
		rules:             rules,
		basicValidator:    NewBasicValidator(rules),
		fieldValidator:    NewFieldValidator(rules),
		businessValidator: NewBusinessValidator(rules),
		securityValidator: NewSecurityValidator(rules),
	}
}

// ValidateConfig performs comprehensive configuration validation
func (cv *ConfigValidator) ValidateConfig(cfg *domain.Config) *ValidationResult {
	start := time.Now()
	result := &ValidationResult{
		IsValid:   true,
		Errors:    []ValidationError{},
		Warnings:  []ValidationWarning{},
		Sanitized: make(map[string]any),
		Timestamp: time.Now(),
	}

	// Level 1: Basic structure validation
	basicResult := cv.basicValidator.ValidateStructure(cfg)
	result.Errors = append(result.Errors, basicResult.Errors...)
	result.Warnings = append(result.Warnings, basicResult.Warnings...)

	// Level 2: Field-level validation with rules
	fieldResult := cv.fieldValidator.ValidateFieldConstraints(cfg)
	result.Errors = append(result.Errors, fieldResult.Errors...)
	result.Warnings = append(result.Warnings, fieldResult.Warnings...)

	// Level 3: Cross-field validation
	cv.validateCrossFieldConstraints(cfg, result)

	// Level 4: Business logic validation
	businessResult := cv.businessValidator.ValidateBusinessLogic(cfg)
	result.Errors = append(result.Errors, businessResult.Errors...)
	result.Warnings = append(result.Warnings, businessResult.Warnings...)

	// Level 5: Security validation
	securityResult := cv.securityValidator.ValidateSecurityConstraints(cfg)
	result.Errors = append(result.Errors, securityResult.Errors...)
	result.Warnings = append(result.Warnings, securityResult.Warnings...)

	// NOTE: Sanitization is NOT applied here to preserve original values
	// Sanitization should be applied separately after validation succeeds
	// This prevents state mutation during verification

	result.Duration = time.Since(start)
	result.IsValid = len(result.Errors) == 0

	return result
}

// ValidateField validates a specific configuration field
func (cv *ConfigValidator) ValidateField(field string, value any) error {
	return cv.fieldValidator.ValidateField(field, value)
}

// validateProfileName validates a profile name
func (cv *ConfigValidator) validateProfileName(profileName string) error {
	if cv.rules.ProfileNamePattern != nil && cv.rules.ProfileNamePattern.Pattern != "" {
		// Basic validation - can be enhanced with regex
		if len(profileName) == 0 {
			return fmt.Errorf("profile name cannot be empty")
		}
		if len(profileName) > 50 {
			return fmt.Errorf("profile name too long")
		}
	}
	return nil
}

// validateCrossFieldConstraints validates constraints between multiple fields
func (cv *ConfigValidator) validateCrossFieldConstraints(cfg *domain.Config, result *ValidationResult) {
	// Check for duplicate protected paths
	if cv.rules.UniquePaths {
		seen := make(map[string]bool)
		for _, path := range cfg.Protected {
			if seen[path] {
				result.AddWarning(
					"protected",
					"Duplicate protected path found",
					"Remove duplicate paths from configuration",
				)
			}
			seen[path] = true
		}
	}

	// Check for duplicate profile names
	if cv.rules.UniqueProfiles {
		seen := make(map[string]bool)
		for name := range cfg.Profiles {
			if seen[name] {
				result.AddError(
					"profiles",
					"duplicate_name",
					name,
					"Duplicate profile name found",
					"Rename one of the profiles",
					SeverityError,
				)
			}
			seen[name] = true
		}
	}

	// Check if safe mode is required but not enabled
	if cv.rules.RequireSafeMode && !cfg.SafeMode {
		result.AddError(
			"safe_mode",
			"required",
			cfg.SafeMode,
			"Safe mode is required by policy",
			"Enable safe mode in configuration",
			SeverityError,
		)
	}

	// Check disk usage vs protected paths balance
	if len(cfg.Protected) > 10 && cfg.MaxDiskUsage > 80 {
		result.AddWarning(
			"max_disk_usage",
			"Consider lowering max disk usage with many protected paths",
			"Balance protection with cleanup effectiveness",
		)
	}
}

// GetRules returns the current validation rules
func (cv *ConfigValidator) GetRules() *ConfigValidationRules {
	return cv.rules
}

// UpdateRules updates the validation rules
func (cv *ConfigValidator) UpdateRules(rules *ConfigValidationRules) {
	cv.rules = rules
	cv.basicValidator = NewBasicValidator(rules)
	cv.fieldValidator = NewFieldValidator(rules)
	cv.businessValidator = NewBusinessValidator(rules)
	cv.securityValidator = NewSecurityValidator(rules)
}
