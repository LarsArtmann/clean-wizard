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

	// Validate generations range
	if settings.Generations < MinNixGenerations || settings.Generations > MaxNixGenerations {
		fieldName := fieldPrefix + ".nix_generations"
		result.Warnings = append(result.Warnings, SanitizationWarning{
			Field:     fieldName,
			Original:  settings.Generations,
			Sanitized: settings.Generations,
			Reason:    fmt.Sprintf("Nix generations must be between %d and %d, got %d", MinNixGenerations, MaxNixGenerations, settings.Generations),
		})
	}

	result.SanitizedFields = append(result.SanitizedFields, fieldPrefix+".nix_generations")
}
