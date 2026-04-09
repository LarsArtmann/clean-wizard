package config

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// Nix generations sanitization constants.
const (
	nixGenerationsMin = 1
	nixGenerationsMax = 10
)

// sanitizeNixGenerationsSettings sanitizes Nix generations settings.
func (cs *ConfigSanitizer) sanitizeNixGenerationsSettings(
	fieldPrefix string, settings *domain.NixGenerationsSettings, result *SanitizationResult,
) {
	if settings == nil {
		return
	}

	// Sanitize generations range (ensure 1-10)
	original := settings.Generations
	if settings.Generations < 1 {
		settings.Generations = 1
		result.addChange(
			fieldPrefix+".generations",
			original,
			settings.Generations,
			"clamped to minimum value",
		)
	} else if settings.Generations > nixGenerationsMax {
		settings.Generations = 10
		result.addChange(
			fieldPrefix+".generations",
			original,
			settings.Generations,
			"clamped to maximum value",
		)
	}

	result.SanitizedFields = append(result.SanitizedFields, fieldPrefix+".nix_generations")
}
