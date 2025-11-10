package config

import (
	"fmt"
	"path/filepath"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// SecurityValidator handles security-related validation
type SecurityValidator struct {
	rules *ConfigValidationRules
}

// NewSecurityValidator creates a new security validator
func NewSecurityValidator(rules *ConfigValidationRules) *SecurityValidator {
	return &SecurityValidator{rules: rules}
}

// ValidateSecurityConstraints validates security-related constraints
func (sv *SecurityValidator) ValidateSecurityConstraints(cfg *domain.Config) *ValidationResult {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
	}

	// Ensure critical system paths are protected
	for _, systemPath := range sv.rules.ProtectedSystemPaths {
		if !sv.isPathProtected(cfg.Protected, systemPath) {
			result.Errors = append(result.Errors, ValidationError{
				Field:      "protected",
				Rule:       "security",
				Value:      cfg.Protected,
				Message:    fmt.Sprintf("Critical system path not protected: %s", systemPath),
				Severity:   SeverityError,
				Suggestion: fmt.Sprintf("Add '%s' to protected paths", systemPath),
				Context: map[string]any{
					"system_path": systemPath,
				},
			})
		}
	}

	// Check for path traversal attacks
	for _, path := range cfg.Protected {
		if sv.isPathTraversal(path) {
			result.Errors = append(result.Errors, ValidationError{
				Field:      "protected",
				Rule:       "security",
				Value:      path,
				Message:    fmt.Sprintf("Potential path traversal detected: %s", path),
				Severity:   SeverityError,
				Suggestion: "Use absolute paths without '..' components",
			})
		}
	}

	// Validate profile security constraints
	for name, profile := range cfg.Profiles {
		if profile.Enabled {
			sv.validateProfileSecurity(name, profile, result)
		}
	}

	return result
}

// isPathProtected checks if a system path is in protected paths
func (sv *SecurityValidator) isPathProtected(protected []string, systemPath string) bool {
	for _, path := range protected {
		if sv.pathMatches(path, systemPath) {
			return true
		}
	}
	return false
}

// pathMatches checks if a protected path matches or contains a system path
func (sv *SecurityValidator) pathMatches(protected, systemPath string) bool {
	// Clean both paths for consistent comparison
	cleanProtected := filepath.Clean(protected)
	cleanSystem := filepath.Clean(systemPath)

	// Exact match
	if cleanProtected == cleanSystem {
		return true
	}

	// Check if protected path is a parent
	return sv.isParentPath(cleanProtected, cleanSystem)
}

// isParentPath checks if parent is a parent of child path
func (sv *SecurityValidator) isParentPath(parent, child string) bool {
	return len(child) > len(parent) &&
		child[:len(parent)] == parent &&
		(parent[len(parent)-1] == '/' || child[len(parent)] == '/')
}

// isPathTraversal checks for path traversal patterns
func (sv *SecurityValidator) isPathTraversal(path string) bool {
	return filepath.Clean(path) != path ||
		len(path) > 0 && path[0] != '/' &&
			(filepath.Clean("../"+path) == filepath.Clean(path) ||
				filepath.Clean(path+"../") != filepath.Clean(path))
}

// validateProfileSecurity validates security constraints for a profile
func (sv *SecurityValidator) validateProfileSecurity(profileName string, profile *domain.Profile, result *ValidationResult) {
	// Check risk level constraints
	if profile.Enabled && profile.MaxRiskLevel.IsHigherThan(sv.rules.MaxRiskLevel) {
		result.Errors = append(result.Errors, ValidationError{
			Field:      fmt.Sprintf("profiles.%s.max_risk_level", profileName),
			Rule:       "security",
			Value:      profile.MaxRiskLevel,
			Message:    fmt.Sprintf("Profile '%s' risk level exceeds maximum allowed", profileName),
			Severity:   SeverityError,
			Suggestion: fmt.Sprintf("Set risk level to %s or lower", sv.rules.MaxRiskLevel),
		})
	}

	// Check operations for security constraints
	for i, op := range profile.Operations {
		if op.Enabled && op.RiskLevel.IsHigherThan(sv.rules.MaxRiskLevel) {
			result.Errors = append(result.Errors, ValidationError{
				Field:      fmt.Sprintf("profiles.%s.operations.%d.risk_level", profileName, i),
				Rule:       "security",
				Value:      op.RiskLevel,
				Message:    fmt.Sprintf("Operation '%s' risk level exceeds maximum allowed", op.Name),
				Severity:   SeverityError,
				Suggestion: fmt.Sprintf("Set operation risk level to %s or lower", sv.rules.MaxRiskLevel),
			})
		}
	}
}
