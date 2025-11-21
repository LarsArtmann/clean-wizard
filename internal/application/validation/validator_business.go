package config

import (
	"fmt"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// validateOperationRisk checks if an operation violates safe mode + critical risk rule
func (cv *ConfigValidator) validateOperationRisk(cfg *domain.Config, profileName string, op domain.CleanupOperation) *ValidationError {
	if op.RiskLevel == domain.RiskCritical && cfg.SafetyLevel == domain.SafetyLevelDisabled {
		return &ValidationError{
			Field:      fmt.Sprintf("profiles.%s.operations.%s.risk_level", profileName, op.Name),
			Rule:       "security",
			Value:      op.RiskLevel,
			Message:    fmt.Sprintf("Critical risk operation '%s' requires safety level enabled", op.Name),
			Severity:   SeverityError,
			Suggestion: "Enable safety level or remove critical risk operations",
		}
	}
	return nil
}

// validateBusinessLogic validates business logic constraints
func (cv *ConfigValidator) validateBusinessLogic(cfg *domain.Config, result *ValidationResult) {
	for name, profile := range cfg.Profiles {
		// Check for nil profile to prevent panic
		if profile == nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:      fmt.Sprintf("profiles.%s", name),
				Rule:       "business_logic",
				Value:      nil,
				Message:    fmt.Sprintf("Profile '%s' is nil", name),
				Severity:   SeverityError,
				Suggestion: "Remove or define profile",
			})
			continue
		}

		// Validate profile business logic
		if len(profile.Operations) == 0 {
			result.Errors = append(result.Errors, ValidationError{
				Field:      fmt.Sprintf("profiles.%s.operations", name),
				Rule:       "business_logic",
				Value:      len(profile.Operations),
				Message:    fmt.Sprintf("Profile '%s' must have at least one operation", name),
				Severity:   SeverityError,
				Suggestion: "Add at least one operation to profile",
			})
			continue
		}

		// Validate each operation
		for _, op := range profile.Operations {
			// Validate risk vs safe mode
			if cfg.SafetyLevel == domain.SafetyLevelDisabled && op.RiskLevel == domain.RiskCritical {
				result.Errors = append(result.Errors, ValidationError{
					Field:      fmt.Sprintf("profiles.%s.operations.%s.risk_level", name, op.Name),
					Rule:       "business_logic",
					Value:      op.RiskLevel,
					Message:    fmt.Sprintf("Critical risk operation '%s' not allowed in unsafe mode", op.Name),
					Severity:   SeverityError,
					Suggestion: "Enable safe mode or remove critical risk operation",
				})
				continue
			}

			// Validate critical risk + safe mode constraint using helper
			if err := cv.validateOperationRisk(cfg, name, op); err != nil {
				result.Errors = append(result.Errors, *err)
			}

			// Validate operation settings
			if op.Settings != nil {
				opType := domain.GetOperationType(op.Name)
				if err := op.Settings.ValidateSettings(opType); err != nil {
					result.Errors = append(result.Errors, ValidationError{
						Field:      fmt.Sprintf("profiles.%s.operations.%s.settings", name, op.Name),
						Rule:       "validation",
						Value:      op.Settings,
						Message:    fmt.Sprintf("Invalid settings for operation '%s': %v", op.Name, err),
						Severity:   SeverityError,
						Suggestion: "Fix operation settings according to validation rules",
					})
				}
			}
		}
	}
}

// validateSecurityConstraints validates security-related constraints
func (cv *ConfigValidator) validateSecurityConstraints(cfg *domain.Config, result *ValidationResult) {
	// Validate no protected paths contain dangerous patterns
	for _, path := range cfg.Protected {
		if path == "/" {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:      "protected",
				Message:    "Protecting root directory '/' may prevent system operations",
				Suggestion: "Consider protecting specific system directories instead",
				Context: &ValidationContext{
					Metadata: map[string]string{
						"protected_path": path,
					},
				},
			})
		}

		// Check for suspicious path patterns
		if strings.Contains(path, "..") {
			result.Errors = append(result.Errors, ValidationError{
				Field:      "protected",
				Rule:       "security",
				Value:      path,
				Message:    "Protected path contains parent directory reference '..'",
				Severity:   SeverityError,
				Suggestion: "Use absolute paths without parent directory references",
			})
		}
	}

	// Validate no profiles have critical operations without explicit consent
	for name, profile := range cfg.Profiles {
		// Check for nil profile to prevent panic
		if profile == nil {
			continue
		}

		for _, op := range profile.Operations {
			if err := cv.validateOperationRisk(cfg, name, op); err != nil {
				result.Errors = append(result.Errors, *err)
			}
		}
	}
}
