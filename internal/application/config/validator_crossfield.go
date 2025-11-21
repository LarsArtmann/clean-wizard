package config

import (
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// validateCrossFieldConstraints validates cross-field relationships
func (cv *ConfigValidator) validateCrossFieldConstraints(cfg *config.Config, result *ValidationResult) {
	// Validate operation count vs risk level
	for name, profile := range cfg.Profiles {
		// Check for nil profile to prevent panic
		if profile == nil {
			continue
		}

		operationCount := len(profile.Operations)
		maxOperations := 100 // Default value

		// Cross-field: operation count vs risk level
		if operationCount > maxOperations {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:      fmt.Sprintf("profiles.%s.operations", name),
				Message:    fmt.Sprintf("Profile '%s' has %d operations (exceeds recommended max of %d)", name, operationCount, maxOperations),
				Suggestion: "Consider splitting profile into smaller, focused profiles",
				Context: &ValidationContext{
					Metadata: map[string]string{
						"operation_count": fmt.Sprintf("%d", operationCount),
						"max_operations":  fmt.Sprintf("%d", maxOperations),
					},
				},
			})
		}
	}

	// Safe mode vs risk level consistency
	if cfg.SafetyLevel == shared.SafetyLevelDisabled {
		maxRisk := shared.RiskLow // Default
		for _, profile := range cfg.Profiles {
			if profile == nil {
				continue
			}
			for _, op := range profile.Operations {
				if op.RiskLevel > maxRisk {
					maxRisk = op.RiskLevel
				}
			}
		}

		if maxRisk == shared.RiskCritical {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:      "safe_mode",
				Message:    "Critical risk operations enabled while safe_mode is false",
				Suggestion: "Enable safe_mode or review critical risk operations",
				Context: &ValidationContext{
					Metadata: map[string]string{
						"max_risk_level": maxRisk.String(),
						"safety_level":   fmt.Sprintf("%v", cfg.SafetyLevel),
					},
				},
			})
		}
	}
}
