package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// sanitizeSystemTempSettings sanitizes system temp settings.
func (cs *ConfigSanitizer) sanitizeSystemTempSettings(fieldPrefix string, settings *domain.SystemTempSettings, result *SanitizationResult) {
	if settings == nil {
		return
	}

	// Sanitize paths array
	if len(settings.Paths) > 0 {
		sanitizedPaths := make([]string, 0, len(settings.Paths))
		for i, path := range settings.Paths {
			original := path

			if cs.rules.TrimWhitespace {
				path = strings.TrimSpace(path)
			}

			if cs.rules.NormalizePaths {
				path = filepath.Clean(path)
			}

			// Validate absolute path requirement
			if !filepath.IsAbs(path) {
				result.Warnings = append(result.Warnings, SanitizationWarning{
					Field:     fmt.Sprintf("%s.paths[%d]", fieldPrefix, i),
					Original:  original,
					Sanitized: path,
					Reason:    "system temp path must be absolute: " + original,
				})
				continue // Skip processing invalid path
			}

			if original != path {
				result.addChange(fmt.Sprintf("%s.paths[%d]", fieldPrefix, i), original, path, "sanitized system temp path")
			}

			sanitizedPaths = append(sanitizedPaths, path)
		}

		// Remove duplicates and sort
		if cs.rules.RemoveDuplicates {
			sanitizedPaths = cs.removeDuplicates(sanitizedPaths)
		}
		if cs.rules.SortArrays {
			cs.sortStrings(sanitizedPaths)
		}

		settings.Paths = sanitizedPaths
	}

	// Sanitize older_than duration
	if cs.rules.TrimWhitespace && settings.OlderThan != "" {
		original := settings.OlderThan
		settings.OlderThan = strings.TrimSpace(settings.OlderThan)
		if original != settings.OlderThan {
			result.addChange(fieldPrefix+".older_than", original, settings.OlderThan, "trimmed whitespace")
		}

		// Validate duration format using custom parser
		if _, err := domain.ParseCustomDuration(settings.OlderThan); err != nil {
			result.Warnings = append(result.Warnings, SanitizationWarning{
				Field:     fieldPrefix + ".older_than",
				Original:  settings.OlderThan,
				Sanitized: settings.OlderThan,
				Reason:    fmt.Sprintf("invalid duration format: %v", err),
			})
		}
	}

	result.SanitizedFields = append(result.SanitizedFields, fieldPrefix+".system_temp")
}
