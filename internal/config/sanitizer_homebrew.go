package config

import (
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// sanitizeHomebrewSettings sanitizes Homebrew settings
func (cs *ConfigSanitizer) sanitizeHomebrewSettings(fieldPrefix string, settings *domain.HomebrewSettings, result *SanitizationResult) {
	if settings == nil {
		return
	}

	// Sanitize prune field
	if cs.rules.TrimWhitespace && settings.Prune != "" {
		original := settings.Prune
		settings.Prune = strings.TrimSpace(settings.Prune)
		if original != settings.Prune {
			result.addChange(fieldPrefix+".prune", original, settings.Prune, "trimmed whitespace")
		}
	}

	result.SanitizedFields = append(result.SanitizedFields, fieldPrefix+".homebrew")
}
