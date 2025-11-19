package config

import (
	"errors"
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// sanitizeOperationSettings sanitizes operation settings with type safety
func (cs *ConfigSanitizer) sanitizeOperationSettings(fieldPrefix, operationName string, settings *domain.OperationSettings, result *SanitizationResult) {
	// Guard against nil settings to prevent panic
	if settings == nil {
		result.Warnings = append(result.Warnings, SanitizationWarning{
			Field:     fieldPrefix,
			Original:  "nil settings",
			Sanitized: "nil settings",
			Reason:    "settings is nil - cannot sanitize",
		})
		return
	}

	opType := domain.GetOperationType(operationName)

	// Validate settings first
	if err := settings.ValidateSettings(opType); err != nil {
		// Convert validation errors to warnings since we're in sanitization
		var vErr *domain.ValidationError
		if errors.As(err, &vErr) {
			result.Warnings = append(result.Warnings, SanitizationWarning{
				Field:     fieldPrefix + "." + vErr.Field,
				Original:  vErr.Value,
				Sanitized: vErr.Value,
				Reason:    vErr.Message,
			})
		} else {
			result.Warnings = append(result.Warnings, SanitizationWarning{
				Field:     fieldPrefix,
				Original:  "settings validation",
				Sanitized: "settings validation",
				Reason:    fmt.Sprintf("validation error: %v", err),
			})
		}
		return
	}

	// Type-aware sanitization based on operation type
	switch opType {
	case domain.OperationTypeNixGenerations:
		cs.sanitizeNixGenerationsSettings(fieldPrefix, settings.NixGenerations, result)

	case domain.OperationTypeTempFiles:
		cs.sanitizeTempFilesSettings(fieldPrefix, settings.TempFiles, result)

	case domain.OperationTypeHomebrew:
		cs.sanitizeHomebrewSettings(fieldPrefix, settings.Homebrew, result)

	case domain.OperationTypeSystemTemp:
		cs.sanitizeSystemTempSettings(fieldPrefix, settings.SystemTemp, result)

	// For custom operation types, no specific sanitization is applied
	// Let validation logic handle invalid operations elsewhere
	}
}
