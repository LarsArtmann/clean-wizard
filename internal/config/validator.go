package config

import (
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// ConfigValidationRules defines all validation constraints
type ConfigValidationRules struct {
	// Numeric Constraints
	MaxDiskUsage      *ValidationRule[int] `json:"max_disk_usage"`
	MinProtectedPaths *ValidationRule[int] `json:"min_protected_paths"`
	MaxProfiles       *ValidationRule[int] `json:"max_profiles"`
	MaxOperations     *ValidationRule[int] `json:"max_operations"`

	// String Constraints
	ProfileNamePattern *ValidationRule[string] `json:"profile_name_pattern"`
	PathPattern        *ValidationRule[string] `json:"path_pattern"`

	// Array Constraints
	UniquePaths    bool `json:"unique_paths"`
	UniqueProfiles bool `json:"unique_profiles"`

	// Safety Constraints
	ProtectedSystemPaths []string `json:"protected_system_paths"`
	RequireSafeMode      bool     `json:"require_safe_mode"`

	// Risk Constraints
	MaxRiskLevel   domain.RiskLevel `json:"max_risk_level"`
	BackupRequired domain.RiskLevel `json:"backup_required"`
}

// ValidationRule represents a validation constraint for a specific type
type ValidationRule[T comparable] struct {
	Required bool   `json:"required"`
	Min      *T     `json:"min,omitempty"`
	Max      *T     `json:"max,omitempty"`
	Pattern  string `json:"pattern,omitempty"`
	Values   []T    `json:"values,omitempty"`
}

// ConfigValidator provides comprehensive type-safe configuration validation
type ConfigValidator struct {
	rules             *ConfigValidationRules
	basicValidator    *BasicValidator
	fieldValidator    *FieldValidator
	businessValidator *BusinessValidator
	securityValidator *SecurityValidator
}

// NewConfigValidator creates a new configuration validator with default rules
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		rules:             getDefaultValidationRules(),
		basicValidator:    NewBasicValidator(),
		fieldValidator:    NewFieldValidator(),
		businessValidator: NewBusinessValidator(),
		securityValidator: NewSecurityValidator(),
	}
}

// NewConfigValidatorWithRules creates a validator with custom rules
func NewConfigValidatorWithRules(rules *ConfigValidationRules) *ConfigValidator {
	return &ConfigValidator{
		rules:             rules,
		basicValidator:    NewBasicValidator(),
		fieldValidator:    NewFieldValidator(),
		businessValidator: NewBusinessValidator(),
		securityValidator: NewSecurityValidator(),
	}
}

// ValidateConfig performs comprehensive configuration validation
func (cv *ConfigValidator) ValidateConfig(cfg *domain.Config) *ValidationResult {
	start := time.Now()
	result := &ValidationResult{
		IsValid:   true,
		Errors:    []ValidationError{},
		Warnings:  []ValidationWarning{},
		Sanitized: &ValidationSanitizedData{
			FieldsModified: []string{},
			RulesApplied:  []string{"basic_validation", "field_validation", "business_validation", "security_validation"},
			Metadata: map[string]string{
				"validation_level": "comprehensive",
				"timestamp":       start.Format(time.RFC3339),
			},
			Data: make(map[string]any),
		},
		Timestamp: start,
	}

	// Basic structure validation
	cv.basicValidator.ValidateBasicStructure(cfg, result)

	// Field-level validation
	cv.fieldValidator.ValidateFields(cfg, cv.rules, result)

	// Business logic validation
	cv.businessValidator.ValidateBusinessRules(cfg, cv.rules, result)

	// Security validation
	cv.securityValidator.ValidateSecurity(cfg, cv.rules, result)

	result.Duration = time.Since(start)
	return result
}

// ValidateProfile performs profile-specific validation
func (cv *ConfigValidator) ValidateProfile(profileName string, profile *domain.Profile) *ValidationResult {
	start := time.Now()
	result := &ValidationResult{
		IsValid:   true,
		Errors:    []ValidationError{},
		Warnings:  []ValidationWarning{},
		Sanitized: &ValidationSanitizedData{
			FieldsModified: []string{},
			RulesApplied:  []string{"profile_validation"},
			Metadata: map[string]string{
				"profile": profileName,
			},
			Data: make(map[string]any),
		},
		Timestamp: start,
	}

	// Validate profile name
	if profile.Name == "" {
		result.AddError("name", "required", "", "Profile name is required", "Set a descriptive name", SeverityError)
	} else if profile.Name != profileName {
		result.AddWarning("name", fmt.Sprintf("Profile name mismatch: config key '%s' vs profile name '%s'", profileName, profile.Name))
	}

	// Validate description
	if profile.Description == "" {
		result.AddWarning("description", "Profile description is recommended for clarity")
	}

	// Validate operations
	if len(profile.Operations) == 0 {
		result.AddError("operations", "required", "", "At least one operation is required", "Add cleanup operations", SeverityError)
	} else {
		for i, op := range profile.Operations {
			cv.validateOperation(&op, fmt.Sprintf("operations[%d]", i), result)
		}
	}

	// Validate risk level
	if profile.MaxRiskLevel != "" && !profile.MaxRiskLevel.IsValid() {
		result.AddError("max_risk_level", "enum", profile.MaxRiskLevel, "Invalid risk level", "Use: LOW, MEDIUM, HIGH, CRITICAL", SeverityError)
	}

	result.Duration = time.Since(start)
	return result
}

// validateOperation validates individual operations
func (cv *ConfigValidator) validateOperation(op *domain.CleanupOperation, fieldPrefix string, result *ValidationResult) {
	// Validate operation name
	if op.Name == "" {
		result.AddError(fieldPrefix+".name", "required", "", "Operation name is required", "Set operation name", SeverityError)
	}

	// Validate description
	if op.Description == "" {
		result.AddWarning(fieldPrefix+".description", "Operation description is recommended")
	}

	// Validate risk level
	if !op.RiskLevel.IsValid() {
		result.AddError(fieldPrefix+".risk_level", "enum", op.RiskLevel, "Invalid risk level", "Use: LOW, MEDIUM, HIGH, CRITICAL", SeverityError)
	}

	// Validate settings
	if op.Settings != nil {
		opType := domain.GetOperationType(op.Name)
		if err := op.Settings.ValidateSettings(opType); err != nil {
			result.AddError(fieldPrefix+".settings", "validation", op.Settings, err.Error(), "Fix operation settings", SeverityError)
		}
	}
}

// getDefaultValidationRules returns default validation rules
func getDefaultValidationRules() *ConfigValidationRules {
	return &ConfigValidationRules{
		MaxDiskUsage: &ValidationRule[int]{
			Required: true,
			Min:      intPtr(1),
			Max:      intPtr(95),
		},
		MinProtectedPaths: &ValidationRule[int]{
			Required: true,
			Min:      intPtr(1),
		},
		MaxProfiles: &ValidationRule[int]{
			Required: false,
			Max:      intPtr(10),
		},
		MaxOperations: &ValidationRule[int]{
			Required: false,
			Max:      intPtr(20),
		},
		ProfileNamePattern: &ValidationRule[string]{
			Required: true,
			Pattern:  "^[a-z][a-z0-9-_]*$",
		},
		PathPattern: &ValidationRule[string]{
			Required: true,
			Pattern:  "^/.*",
		},
		UniquePaths:    true,
		UniqueProfiles: true,
		ProtectedSystemPaths: []string{
			"/", "/System", "/Library", "/usr", "/etc", "/bin", "/sbin",
		},
		RequireSafeMode: true,
		MaxRiskLevel:   domain.RiskCritical,
		BackupRequired: domain.RiskHigh,
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}