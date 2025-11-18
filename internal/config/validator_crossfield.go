package config

import (
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// validateFieldConstraints validates individual fields against rules
func (cv *ConfigValidator) validateFieldConstraints(cfg *domain.Config, result *ValidationResult) {
	// MaxDiskUsage validation
	if err := cv.validateMaxDiskUsage(cfg.MaxDiskUsage); err != nil {
		minUsage, maxUsage := cv.getMaxDiskUsageBounds()
		
		// Create descriptive suggestion using actual rule values
		suggestion := fmt.Sprintf("Set max_disk_usage between %d and %d", minUsage, maxUsage)
		if cv.rules.MaxDiskUsage != nil && cv.rules.MaxDiskUsage.Message != "" {
			suggestion = cv.rules.MaxDiskUsage.Message
		}
		
		result.Errors = append(result.Errors, ValidationError{
			Field:      "max_disk_usage",
			Rule:       "range",
			Value:      cfg.MaxDiskUsage,
			Message:    err.Error(),
			Severity:   SeverityError,
			Suggestion: suggestion,
			Context: map[string]any{
				"min": minUsage,
				"max": maxUsage,
			},
		})
	}

	// Protected paths validation
	if err := cv.validateProtectedPaths(cfg.Protected); err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:      "protected",
			Rule:       "format",
			Value:      cfg.Protected,
			Message:    err.Error(),
			Severity:   SeverityError,
			Suggestion: "Ensure all paths are valid absolute paths",
		})
	}

	// Check for duplicate paths
	if cv.rules.UniquePaths {
		duplicates := cv.findDuplicatePaths(cfg.Protected)
		if len(duplicates) > 0 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:      "protected",
				Message:    fmt.Sprintf("Duplicate protected paths found: %v", duplicates),
				Suggestion: "Remove duplicate paths from protected list",
			})
		}
	}

	// Check profile count limits
	if cv.rules.MaxProfiles != nil && cv.rules.MaxProfiles.Max != nil && len(cfg.Profiles) > *cv.rules.MaxProfiles.Max {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      "profiles",
			Message:    fmt.Sprintf("Profile count (%d) exceeds recommended limit (%d)", len(cfg.Profiles), *cv.rules.MaxProfiles.Max),
			Suggestion: "Consider consolidating profiles to improve maintainability",
		})
	}
}

// validateCrossFieldConstraints validates relationships between fields
func (cv *ConfigValidator) validateCrossFieldConstraints(cfg *domain.Config, result *ValidationResult) {
	// Safe mode vs risk level consistency
	if !cfg.SafeMode {
		maxRisk := cv.findMaxRiskLevel(cfg)
		if maxRisk == domain.RiskCritical {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:      "safe_mode",
				Message:    "Critical risk operations enabled while safe_mode is false",
				Suggestion: "Enable safe_mode or review critical risk operations",
				Context: map[string]any{
					"max_risk_level": maxRisk,
					"safe_mode":      cfg.SafeMode,
				},
			})
		}
	}

	// Validate operation count per profile
	for name, profile := range cfg.Profiles {
		if cv.rules.MaxOperations != nil && cv.rules.MaxOperations.Max != nil {
			if len(profile.Operations) > *cv.rules.MaxOperations.Max {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Field:      fmt.Sprintf("profiles.%s.operations", name),
					Message:    fmt.Sprintf("Profile '%s' has %d operations, exceeding recommended limit (%d)", name, len(profile.Operations), *cv.rules.MaxOperations.Max),
					Suggestion: "Consider splitting operations into multiple profiles",
					Context: map[string]any{
						"operation_count": len(profile.Operations),
						"max_operations":  *cv.rules.MaxOperations.Max,
					},
				})
			}
		}
	}
}