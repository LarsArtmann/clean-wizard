package config

import (
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// equalStringSlices helper to compare string slices
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// sanitizeHomebrewSettings sanitizes Homebrew settings
func (cs *ConfigSanitizer) sanitizeHomebrewSettings(fieldPrefix string, settings *domain.HomebrewSettings, result *SanitizationResult) {
	if settings == nil {
		return
	}

	// Guard against nil rules - use sane defaults if rules are missing
	if cs.rules == nil {
		return
	}

	originalSettings := *settings
	
	// Sanitize prune field
	if cs.rules.TrimWhitespace && settings.Prune != "" {
		original := settings.Prune
		settings.Prune = strings.TrimSpace(settings.Prune)
		if original != settings.Prune {
			result.addChange(fieldPrefix+".prune", original, settings.Prune, "trimmed whitespace")
		}
	}

	// Only mark as sanitized if changes actually occurred
	if originalSettings.Prune != settings.Prune {
		result.SanitizedFields = append(result.SanitizedFields, fieldPrefix+".homebrew")
	}
}
