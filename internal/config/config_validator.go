package config

import (
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// ConfigValidator provides comprehensive type-safe configuration validation
// This uses simplified validation approach for maintainability
type ConfigValidator struct {
	rules *ConfigValidationRules
}

// NewConfigValidator creates a new configuration validator with default rules
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		rules: getDefaultValidationRules(),
	}
}

// NewConfigValidatorWithRules creates a validator with custom rules
func NewConfigValidatorWithRules(rules *ConfigValidationRules) *ConfigValidator {
	return &ConfigValidator{
		rules: rules,
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
			RulesApplied:  []string{"comprehensive_validation"},
			Metadata: map[string]string{
				"validation_level": "comprehensive",
				"timestamp":       start.Format(time.RFC3339),
			},
			Data: make(map[string]any),
		},
		Timestamp: start,
	}

	// Basic structure validation
	if cfg == nil {
		result.AddError("config", "required", nil, "Configuration cannot be nil", "Provide valid configuration", SeverityCritical)
		return result
	}

	// Validate version
	if cfg.Version == "" {
		result.AddError("version", "required", "", "Version is required", "Set configuration version", SeverityError)
	}

	// Validate MaxDiskUsage using direct logic
	if cfg.MaxDiskUsage < 1 || cfg.MaxDiskUsage > 95 {
		result.AddError("max_disk_usage", "range", cfg.MaxDiskUsage, 
			"MaxDiskUsage must be between 1 and 95", 
			"Set MaxDiskUsage to valid range", SeverityError)
	}

	// Validate protected paths
	if len(cfg.Protected) == 0 {
		result.AddError("protected", "required", "", "Protected paths cannot be empty", "Add system paths to protect", SeverityError)
	}

	// Validate profiles
	if len(cfg.Profiles) == 0 {
		result.AddError("profiles", "required", "", "At least one profile is required", "Add cleanup profiles", SeverityError)
	}

	// Validate each profile
	for name, profile := range cfg.Profiles {
		if profile.Name == "" {
			result.AddError(fmt.Sprintf("profiles.%s.name", name), "required", "", "Profile name is required", "Set profile name", SeverityError)
		}
		if len(profile.Operations) == 0 {
			result.AddError(fmt.Sprintf("profiles.%s.operations", name), "required", "", "At least one operation is required", "Add cleanup operations", SeverityError)
		}
		
		// Validate operations
		for i, op := range profile.Operations {
			fieldPrefix := fmt.Sprintf("profiles.%s.operations[%d]", name, i)
			if op.Name == "" {
				result.AddError(fieldPrefix+".name", "required", "", "Operation name is required", "Set operation name", SeverityError)
			}
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
	}

	result.Duration = time.Since(start)
	return result
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