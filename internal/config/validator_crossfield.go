package config

import (
	"fmt"
	"strings"

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
			Context: &ValidationContext{
				MinValue: minUsage,
				MaxValue: maxUsage,
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
	// Safe mode vs risk level consistency - only if no per-operation critical risk errors
	if !cfg.SafeMode {
		maxRisk := cv.findMaxRiskLevel(cfg)
		if maxRisk == domain.RiskCritical {
			// Check if we already have per-operation critical risk errors
			hasCriticalRiskErrors := false
			for _, err := range result.Errors {
				if strings.Contains(err.Field, "risk_level") && strings.Contains(err.Message, "Critical risk operation") {
					hasCriticalRiskErrors = true
					break
				}
			}
			
			// Only add warning if no per-operation errors exist
			if !hasCriticalRiskErrors {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Field:      "safe_mode",
					Message:    "Critical risk operations enabled while safe_mode is false",
					Suggestion: "Enable safe_mode or review critical risk operations",
					Context: &ValidationContext{
						Metadata: map[string]string{
							"max_risk_level": maxRisk.String(),
							"safe_mode":      fmt.Sprintf("%v", cfg.SafeMode),
						},
					},
				})
			}
=======
	// Safe mode vs risk level consistency
	if !cfg.SafeMode {
		maxRisk := cv.findMaxRiskLevel(cfg)
		if maxRisk == domain.RiskCritical {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:      "safe_mode",
				Message:    "Critical risk operations enabled while safe_mode is false",
				Suggestion: "Enable safe_mode or review critical risk operations",
				Context: &ValidationContext{
					Metadata: map[string]string{
						"max_risk_level": maxRisk.String(),
						"safe_mode":      fmt.Sprintf("%v", cfg.SafeMode),
					},
				},
			})
>>>>>>> master
		}
	}

	// Validate operation count per profile
	for name, profile := range cfg.Profiles {
		// Check for nil profile to prevent panic
		if profile == nil {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:      fmt.Sprintf("profiles.%s", name),
				Message:    fmt.Sprintf("Profile '%s' is nil", name),
				Suggestion: "Remove or define the profile",
			})
			continue
		}

		if cv.rules.MaxOperations != nil && cv.rules.MaxOperations.Max != nil {
			if len(profile.Operations) > *cv.rules.MaxOperations.Max {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Field:      fmt.Sprintf("profiles.%s.operations", name),
					Message:    fmt.Sprintf("Profile '%s' has %d operations, exceeding recommended limit (%d)", name, len(profile.Operations), *cv.rules.MaxOperations.Max),
					Suggestion: "Consider splitting operations into multiple profiles",
					Context: &ValidationContext{
						Metadata: map[string]string{
							"operation_count": fmt.Sprintf("%d", len(profile.Operations)),
							"max_operations":  fmt.Sprintf("%d", *cv.rules.MaxOperations.Max),
						},
					},
				})
			}
		}
	}
}
