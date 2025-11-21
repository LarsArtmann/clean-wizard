package config

import (
	"fmt"
	"path/filepath"

	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// validateFieldConstraints validates individual fields against rules
func (cv *ConfigValidator) validateFieldConstraints(cfg *domain.Config, result *ValidationResult) {
	// Validate max disk usage
	if cfg.MaxDiskUsage < 0 || cfg.MaxDiskUsage > 100 {
		result.Errors = append(result.Errors, ValidationError{
			Field:      "max_disk_usage",
			Rule:       "range",
			Value:      cfg.MaxDiskUsage,
			Message:    "Max disk usage must be between 0 and 100 percent",
			Severity:   SeverityError,
			Suggestion: "Set a reasonable disk usage percentage (e.g., 80)",
		})
	}

	// Validate protected paths
	for i, path := range cfg.Protected {
		if path == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:      fmt.Sprintf("protected[%d]", i),
				Rule:       "required",
				Value:      path,
				Message:    "Protected path cannot be empty",
				Severity:   SeverityError,
				Suggestion: "Provide a valid absolute path",
			})
		} else if !filepath.IsAbs(path) {
			result.Errors = append(result.Errors, ValidationError{
				Field:      fmt.Sprintf("protected[%d]", i),
				Rule:       "format",
				Value:      path,
				Message:    "Protected path must be absolute",
				Severity:   SeverityError,
				Suggestion: "Use absolute path (e.g., /System, /Applications)",
			})
		}
	}

	// Check for duplicate protected paths
	duplicates := cv.findDuplicatePaths(cfg.Protected)
	for _, path := range duplicates {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      "protected",
			Message:    fmt.Sprintf("Duplicate protected path: %s", path),
			Suggestion: "Remove duplicate entries from protected paths",
			Context: &ValidationContext{
				Metadata: map[string]string{
					"duplicate_path": path,
				},
			},
		})
	}
}

// findDuplicatePaths finds duplicate paths in given slice
func (cv *ConfigValidator) findDuplicatePaths(paths []string) []string {
	seen := make(map[string]bool)
	duplicates := []string{}

	for _, path := range paths {
		if seen[path] {
			duplicates = append(duplicates, path)
		} else {
			seen[path] = true
		}
	}

	return duplicates
}
