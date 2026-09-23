package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
)

// sanitizeTempFilesSettings sanitizes temporary files settings.
func (cs *ConfigSanitizer) sanitizeTempFilesSettings(
	fieldPrefix string, settings *operations.TempFilesSettings, result *SanitizationResult,
) {
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
			exclude = cs.sanitizePathValue(exclude)

			if original != exclude {
				result.addChange(
					fmt.Sprintf("%s.excludes[%d]", fieldPrefix, i),
					original,
					exclude,
					"sanitized exclude path",
				)
			}

			sanitizedExcludes = append(sanitizedExcludes, exclude)
		}

		settings.Excludes = cs.finalizePathList(sanitizedExcludes)
	}

	result.SanitizedFields = append(result.SanitizedFields, fieldPrefix+".temp_files")
}
