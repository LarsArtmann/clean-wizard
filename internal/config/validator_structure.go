package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// Precompiled semver regex for performance optimization
var semverRegex = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

// isValidSemver validates semantic version format (MAJOR.MINOR.PATCH)
func isValidSemver(version string) bool {
	if version == "" {
		return false
	}
	
	// Use precompiled regex for better performance
	return semverRegex.MatchString(version)
}

// validateBasicStructure validates basic configuration structure
func (cv *ConfigValidator) validateBasicStructure(cfg *domain.Config, result *ValidationResult) {
	// Version validation
	if cfg.Version == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:      "version",
			Rule:       "required",
			Value:      cfg.Version,
			Message:    "Configuration version is required",
			Severity:   SeverityError,
			Suggestion: "Add version field with semantic version (e.g., '1.0.0')",
		})
	} else if !isValidSemver(cfg.Version) {
		result.Errors = append(result.Errors, ValidationError{
			Field:      "version",
			Rule:       "semver_format",
			Value:      cfg.Version,
			Message:    "Version must follow semantic version format (MAJOR.MINOR.PATCH)",
			Severity:   SeverityError,
			Suggestion: "Change version to valid semantic version (e.g., '1.0.0', '2.1.3', '1.0.0-alpha')",
		})
	}

	// Profiles validation
	if len(cfg.Profiles) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:      "profiles",
			Rule:       "required",
			Value:      cfg.Profiles,
			Message:    "At least one profile is required",
			Severity:   SeverityError,
			Suggestion: "Add a profile with at least one operation",
		})
	}

	// Protected paths validation
	if len(cfg.Protected) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:      "protected",
			Rule:       "required",
			Value:      cfg.Protected,
			Message:    "Protected paths cannot be empty",
			Severity:   SeverityError,
			Suggestion: "Add system paths like /System, /Applications, /Library",
		})
	} else {
		// Validate each individual protected path
		for i, path := range cfg.Protected {
			trimmedPath := strings.TrimSpace(path)
			fieldName := fmt.Sprintf("protected[%d]", i)

			// Check for empty or whitespace-only paths
			if trimmedPath == "" {
				result.Errors = append(result.Errors, ValidationError{
					Field:      fieldName,
					Rule:       "required",
					Value:      path,
					Message:    "Protected path cannot be empty or whitespace only",
					Severity:   SeverityError,
					Suggestion: "Provide a valid absolute path (e.g., '/System', '/Applications')",
				})
				continue
			}

			// Validate basic path format (must start with "/")
			if !strings.HasPrefix(trimmedPath, "/") {
				result.Errors = append(result.Errors, ValidationError{
					Field:      fieldName,
					Rule:       "path_format",
					Value:      path,
					Message:    "Protected path must be an absolute path starting with '/'",
					Severity:   SeverityError,
					Suggestion: fmt.Sprintf("Change '%s' to an absolute path (e.g., '/%s')", path, strings.TrimLeft(path, "/")),
				})
			}
		}
	}
}
