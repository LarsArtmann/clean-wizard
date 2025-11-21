package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// sanitizeTempFilesSettings sanitizes temporary files settings
func (cs *config.ConfigSanitizer) sanitizeTempFilesSettings(fieldPrefix string, settings *shared.TempFilesSettings, result *SanitizationResult) {
	if settings == nil {
		return
	}

	// Sanitize older_than duration
	originalOlderThan := settings.OlderThan

	if cs.rules.TrimWhitespace && settings.OlderThan != "" {
		settings.OlderThan = strings.TrimSpace(settings.OlderThan)
		if originalOlderThan != settings.OlderThan {
			result.addChange(fieldPrefix+".older_than", originalOlderThan, settings.OlderThan, "trimmed whitespace")
		}
	}

	// Always validate duration format if not empty (regardless of TrimWhitespace flag)
	if settings.OlderThan != "" {
		if parsedDuration, err := shared.ParseCustomDuration(settings.OlderThan); err != nil {
			result.Warnings = append(result.Warnings, SanitizationWarning{
				Field:     fieldPrefix + ".older_than",
				Original:  originalOlderThan, // Use original pre-trim value
				Sanitized: settings.OlderThan,
				Reason:    fmt.Sprintf("invalid duration format: %v", err),
			})
		} else {
			// Normalize the duration to canonical form
			normalizedDuration := shared.FormatDuration(parsedDuration)
			if strings.TrimSpace(originalOlderThan) != normalizedDuration {
				settings.OlderThan = normalizedDuration
				result.addChange(fieldPrefix+".older_than", originalOlderThan, normalizedDuration, "normalized duration")
			}
		}
	}

	// Sanitize excludes array
	if len(settings.Excludes) > 0 {
		sanitizedExcludes := make([]string, 0, len(settings.Excludes))
		for i, exclude := range settings.Excludes {
			original := exclude

			if cs.rules.TrimWhitespace {
				exclude = strings.TrimSpace(exclude)
			}

			if cs.rules.NormalizePaths {
				exclude = filepath.Clean(exclude)
			}

			// Skip empty entries or entries that normalize to "." to avoid excluding current directory
			if exclude == "" || exclude == "." {
				result.addChange(fmt.Sprintf("%s.excludes[%d]", fieldPrefix, i), original, "", "removed empty or current directory reference")
				continue
			}

			if original != exclude {
				result.addChange(fmt.Sprintf("%s.excludes[%d]", fieldPrefix, i), original, exclude, "sanitized exclude path")
			}

			sanitizedExcludes = append(sanitizedExcludes, exclude)
		}

		// Remove duplicates and sort
		if cs.rules.RemoveDuplicates {
			sanitizedExcludes = cs.removeDuplicates(sanitizedExcludes)
		}
		if cs.rules.SortArrays {
			cs.sortStrings(sanitizedExcludes)
		}

		settings.Excludes = sanitizedExcludes
	}

	result.SanitizedFields = append(result.SanitizedFields, fieldPrefix+".temp_files")
}
