package config

import (
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// sanitizeOperationSettings sanitizes operation settings with type safety
func (cs *ConfigSanitizer) sanitizeOperationSettings(fieldPrefix string, operationName string, settings *domain.OperationSettings, result *SanitizationResult) {
	opType := domain.GetOperationType(operationName)
	
	// Validate settings first
	if err := settings.ValidateSettings(opType); err != nil {
		// Convert validation errors to warnings since the result type doesn't have an Errors field
		if validationErr, ok := err.(*domain.ValidationError); ok {
			result.Warnings = append(result.Warnings, SanitizationWarning{
				Field:     fieldPrefix + "." + validationErr.Field,
				Original:  validationErr.Value,
				Sanitized: validationErr.Value,
				Reason:    validationErr.Message,
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
		
	default:
		// For custom operation types, just record that they were processed
		result.Warnings = append(result.Warnings, SanitizationWarning{
			Field:     fieldPrefix,
			Original:  "custom operation settings",
			Sanitized: "custom operation settings",
			Reason:    "custom operation type - no specific sanitization applied",
		})
	}
}