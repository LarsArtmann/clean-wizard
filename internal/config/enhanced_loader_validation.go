package config

import (
	"fmt"
	"slices"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// ConfigSchema represents the configuration schema
type ConfigSchema struct {
	Version     string                 `json:"version"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Types       map[string]SchemaType  `json:"types"`
	Validation  *ConfigValidationRules `json:"validation"`
}

// SchemaType represents a type definition in the schema
type SchemaType struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Required    bool                   `json:"required"`
	Properties  map[string]*SchemaType `json:"properties,omitempty"`
	Items       *SchemaType            `json:"items,omitempty"`
	Enum        []any                  `json:"enum,omitempty"`
	Pattern     string                 `json:"pattern,omitempty"`
	Minimum     *float64               `json:"minimum,omitempty"`
	Maximum     *float64               `json:"maximum,omitempty"`
}

// GetConfigSchema returns the configuration schema for validation
func (ecl *EnhancedConfigLoader) GetConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Version:     "1.0.0",
		Title:       "Clean Wizard Configuration Schema",
		Description: "Comprehensive configuration schema for clean-wizard",
		Types:       ecl.generateSchemaTypes(),
		Validation:  ecl.validator.rules,
	}
}

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
	requiredPaths := []string{"/System", "/Library"}
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

// generateSchemaTypes generates type definitions for the schema
func (ecl *EnhancedConfigLoader) generateSchemaTypes() map[string]SchemaType {
	return map[string]SchemaType{
		"Config": {
			Type:        "object",
			Description: "Main configuration structure",
			Required:    true,
			Properties: map[string]*SchemaType{
				"version": {
					Type:        "string",
					Description: "Configuration version",
					Required:    true,
					Pattern:     "^\\d+\\.\\d+\\.\\d+$",
				},
				"safe_mode": {
					Type:        "boolean",
					Description: "Enable safe mode",
					Required:    true,
				},
				"max_disk_usage": {
					Type:        "integer",
					Description: "Maximum disk usage percentage",
					Required:    true,
					Minimum:     func() *float64 { v := 10.0; return &v }(),
					Maximum:     func() *float64 { v := 95.0; return &v }(),
				},
				"protected": {
					Type:        "array",
					Description: "Protected paths",
					Required:    true,
					Items: &SchemaType{
						Type:    "string",
						Pattern: "^/.*",
					},
				},
				"profiles": {
					Type:        "object",
					Description: "Cleaning profiles",
					Required:    true,
				},
			},
		},
	}
}
