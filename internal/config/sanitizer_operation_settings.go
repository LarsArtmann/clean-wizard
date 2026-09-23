package config

import (
	"errors"
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
)

// sanitizeOperationSettings sanitizes operation settings with type safety.
func (cs *ConfigSanitizer) sanitizeOperationSettings(
	fieldPrefix, operationName string,
	settings *operations.OperationSettings, result *SanitizationResult,
) {
	opType := operations.GetOperationType(operationName)

	// Validate settings first
	err := settings.ValidateSettings(opType)
	if err != nil {
		// Convert validation errors to warnings since the result type doesn't have an Errors field
		validationErr := &operations.ValidationError{} //nolint:exhaustruct
		if errors.As(err, &validationErr) {
			result.Warnings = append(result.Warnings, SanitizationWarning{ //nolint:exhaustruct
				Field:     fieldPrefix + "." + validationErr.Field,
				Original:  validationErr.Value,
				Sanitized: validationErr.Value,
				Reason:    validationErr.Message,
			})
		} else {
			result.Warnings = append(result.Warnings, SanitizationWarning{ //nolint:exhaustruct
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
	case operations.OperationTypeNixGenerations:
		cs.sanitizeNixGenerationsSettings(fieldPrefix, settings.NixGenerations, result)

	case operations.OperationTypeTempFiles:
		cs.sanitizeTempFilesSettings(fieldPrefix, settings.TempFiles, result)

	case operations.OperationTypeHomebrew:
		cs.sanitizeHomebrewSettings(fieldPrefix, settings.Homebrew, result)

	case operations.OperationTypeSystemTemp:
		cs.sanitizeSystemTempSettings(fieldPrefix, settings.SystemTemp, result)

	case operations.OperationTypeNodePackages,
		operations.OperationTypeGoPackages,
		operations.OperationTypeCargoPackages,
		operations.OperationTypeBuildCache,
		operations.OperationTypeDocker,
		operations.OperationTypeSystemCache,
		operations.OperationTypeProjectsManagementAutomation,
		operations.OperationTypeProjectExecutables,
		operations.OperationTypeCompiledBinaries,
		operations.OperationTypeGitHistory,
		operations.OperationTypeGolangciLintCache:
		// These operation types have no specific sanitization logic yet
		// Fall through to default handling

	default:
		// For custom operation types, just record that they were processed
		result.Warnings = append(result.Warnings, SanitizationWarning{ //nolint:exhaustruct
			Field:     fieldPrefix,
			Original:  "custom operation settings",
			Sanitized: "custom operation settings",
			Reason:    "custom operation type - no specific sanitization applied",
		})
	}
}
