package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// sanitizeTempFilesSettings sanitizes temporary files settings.
func (cs *ConfigSanitizer) sanitizeTempFilesSettings(fieldPrefix string, settings *domain.TempFilesSettings, result *SanitizationResult) {
	if settings == nil {
		return
	}

	// Sanitize older_than duration
	cs.sanitizeOlderThan(fieldPrefix, &settings.OlderThan, result)

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
