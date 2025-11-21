package config

import (
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// Constants for Nix generations validation
const (
	MinNixGenerations = 1
	MaxNixGenerations = 1000
)

// sanitizeNixGenerationsSettings sanitizes Nix generations settings
func (cs *ConfigSanitizer) sanitizeNixGenerationsSettings(fieldPrefix string, settings *domain.NixGenerationsSettings, result *SanitizationResult) {
	if settings == nil {
		return
	}

	changed := false
	original := settings.Generations
	fieldName := fieldPrefix + ".nix_generations"

	// Validate and clamp generations range
	if settings.Generations < MinNixGenerations {
		settings.Generations = MinNixGenerations
		result.Warnings = append(result.Warnings, SanitizationWarning{
			Field:     fieldName,
			Original:  original,
			Sanitized: settings.Generations,
			Reason:    fmt.Sprintf("Nix generations clamped to minimum %d", MinNixGenerations),
		})
		result.addChange(fieldName, original, settings.Generations, "clamped to minimum")
		changed = true
	} else if settings.Generations > MaxNixGenerations {
		settings.Generations = MaxNixGenerations
		result.Warnings = append(result.Warnings, SanitizationWarning{
			Field:     fieldName,
			Original:  original,
			Sanitized: settings.Generations,
			Reason:    fmt.Sprintf("Nix generations clamped to maximum %d", MaxNixGenerations),
		})
		result.addChange(fieldName, original, settings.Generations, "clamped to maximum")
		changed = true
	}

	// Only mark as sanitized if changes actually occurred
	if changed || original != settings.Generations {
		result.SanitizedFields = append(result.SanitizedFields, fieldPrefix+".nix_generations")
	}
}
