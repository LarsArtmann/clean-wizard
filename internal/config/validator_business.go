package config

import (
	"fmt"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// validateBusinessLogic validates business logic constraints.
func (cv *ConfigValidator) validateBusinessLogic(cfg *domain.Config, result *ValidationResult) {
	for name, profile := range cfg.Profiles {
		// Validate profile business logic
		if len(profile.Operations) == 0 {
			result.Errors = append(result.Errors, ValidationError{
				Field:      fmt.Sprintf("profiles.%s.operations", name),
				Rule:       "business_logic",
				Value:      len(profile.Operations),
				Message:    fmt.Sprintf("Profile '%s' must have at least one operation", name),
				Severity:   SeverityError,
				Suggestion: "Add at least one valid operation to profile",
			})
		}

		// Validate operations within profile
		for _, op := range profile.Operations {
			// Validate risk vs safe mode
			if !cfg.SafeMode.IsEnabled() && op.RiskLevel == domain.RiskLevelType(domain.RiskLevelCriticalType) {
				result.Errors = append(result.Errors, ValidationError{
					Field:      fmt.Sprintf("profiles.%s.operations.%s.risk_level", name, op.Name),
					Rule:       "business_logic",
					Value:      op.RiskLevel,
					Message:    fmt.Sprintf("Critical risk operation '%s' not allowed in unsafe mode", op.Name),
					Severity:   SeverityError,
					Suggestion: "Enable safe mode or remove critical risk operation",
				})
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

	// Validate protected paths don't conflict with operation targets
	for _, profile := range cfg.Profiles {
		for _, op := range profile.Operations {
			if err := cv.validateProtectedPathsConflict(cfg.Protected, op); err != nil {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Field:      fmt.Sprintf("profiles.%s.operations.%s", profile.Name, op.Name),
					Message:    fmt.Sprintf("Operation may affect protected paths: %v", err),
					Suggestion: "Review operation scope and protected paths configuration",
				})
			}
		}
	}
}

// validateSecurityConstraints validates security-related constraints.
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
		for _, op := range profile.Operations {
			if op.RiskLevel == domain.RiskLevelType(domain.RiskLevelCriticalType) && !cfg.SafeMode.IsEnabled() {
				result.Errors = append(result.Errors, ValidationError{
					Field:      fmt.Sprintf("profiles.%s.operations.%s.risk_level", name, op.Name),
					Rule:       "security",
					Value:      op.RiskLevel,
					Message:    fmt.Sprintf("Critical risk operation '%s' requires safe mode enabled", op.Name),
					Severity:   SeverityError,
					Suggestion: "Enable safe mode or remove critical risk operations",
				})
			}
		}
	}
}

// validateProtectedPathsConflict checks if operations might affect protected paths.
func (cv *ConfigValidator) validateProtectedPathsConflict(protected []string, op domain.CleanupOperation) error {
	switch op.Name {
	case "temp-files":
		// Check if temp files cleanup might affect protected paths
		return cv.checkTempFilesConflict(protected, op)
	case "nix-generations":
		// Check if nix cleanup might affect protected paths
		return cv.checkNixConflict(protected, op)
	default:
		// Generic conflict check
		return nil // Skip unknown operations
	}
}

// checkTempFilesConflict checks for temp files conflicts.
func (cv *ConfigValidator) checkTempFilesConflict(protected []string, op domain.CleanupOperation) error {
	if op.Settings != nil && op.Settings.TempFiles != nil {
		for _, exclude := range op.Settings.TempFiles.Excludes {
			for _, protectedPath := range protected {
				if strings.HasPrefix(exclude, protectedPath) || strings.HasPrefix(protectedPath, exclude) {
					return fmt.Errorf("temp files exclude '%s' conflicts with protected path '%s'", exclude, protectedPath)
				}
			}
		}
	}
	return nil
}

// checkNixConflict checks for Nix conflicts.
func (cv *ConfigValidator) checkNixConflict(protected []string, op domain.CleanupOperation) error {
	// Nix operations typically affect /nix/store, check if protected paths overlap
	nixStorePath := "/nix/store"
	for _, protectedPath := range protected {
		if strings.HasPrefix(protectedPath, nixStorePath) || strings.HasPrefix(nixStorePath, protectedPath) {
			return fmt.Errorf("nix operations may conflict with protected path '%s'", protectedPath)
		}
	}
	return nil
}

// findMaxRiskLevel finds the maximum risk level in configuration.
func (cv *ConfigValidator) findMaxRiskLevel(cfg *domain.Config) domain.RiskLevelType {
	maxRisk := domain.RiskLevelType(domain.RiskLevelLowType)
	for _, profile := range cfg.Profiles {
		maxRisk = maxRiskLevelFromOperations(profile.Operations, maxRisk)
		if maxRisk == domain.RiskLevelType(domain.RiskLevelCriticalType) {
			return domain.RiskLevelType(domain.RiskLevelCriticalType)
		}
	}
	return maxRisk
}

func (cv *ConfigValidator) isPathProtected(protected []string, targetPath string) bool {
	for _, path := range protected {
		if strings.HasPrefix(path, targetPath) || strings.HasPrefix(targetPath, path) {
			return true
		}
	}
	return false
}
