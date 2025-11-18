package config

import (
	"fmt"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// sanitizeNixGenerationsSettings sanitizes Nix generations settings
func (cs *ConfigSanitizer) sanitizeNixGenerationsSettings(fieldPrefix string, settings *domain.NixGenerationsSettings, result *SanitizationResult) {
	if settings == nil {
		return
	}

	// Validate generations range (ensure 1-1000)
	if settings.Generations < 1 || settings.Generations > 1000 {
		result.Warnings = append(result.Warnings, SanitizationWarning{
			Field:     fieldPrefix + ".generations",
			Original:  settings.Generations,
			Sanitized: settings.Generations,
			Reason:    fmt.Sprintf("Nix generations must be between 1 and 1000, got %d", settings.Generations),
		})
	}

	result.SanitizedFields = append(result.SanitizedFields, fieldPrefix+".nix_generations")
}
