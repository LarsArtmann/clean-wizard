package config

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
)

// BDDTestHelpers provides standardized utilities for BDD testing
// Eliminates duplicate patterns across test files

// ConfigModifier represents a function that modifies a config.
type ConfigModifier func(*types.Config) *types.Config

// ProfileOperationModifier modifies specific profile operations.
type ProfileOperationModifier func(*types.Profile, *types.CleanupOperation) bool // returns true if modified

// FindProfileOperation finds a specific operation within a profile by name
// Returns the operation index, or -1 if not found.
func FindProfileOperation(
	cfg *types.Config,
	profileName, operationName string,
) (*types.Profile, int) {
	if cfg == nil {
		return nil, -1
	}

	profile, exists := cfg.Profiles[profileName]
	if !exists {
		return nil, -1
	}

	for i, op := range profile.Operations {
		if op.Name == operationName {
			return profile, i
		}
	}

	return profile, -1
}

// ModifyProfileOperation applies a modifier to a specific operation.
// Returns true if the operation was found and modified.
func ModifyProfileOperation(
	cfg *types.Config, profileName, operationName string,
	modifier ProfileOperationModifier,
) bool {
	profile, opIndex := FindProfileOperation(cfg, profileName, operationName)
	if profile == nil || opIndex == -1 {
		return false
	}

	return modifier(profile, &profile.Operations[opIndex])
}

// WithOperationSettings applies a settings modifier to a specific operation.
// Returns true if the operation was found and modified.
func WithOperationSettings(
	cfg *types.Config,
	profileName, operationName string,
	settingsModifier func(*operations.OperationSettings) bool,
) bool {
	return ModifyProfileOperation(
		cfg,
		profileName,
		operationName,
		func(profile *types.Profile, op *types.CleanupOperation) bool {
			if op.Settings == nil {
				return false
			}

			return settingsModifier(op.Settings)
		},
	)
}

// BoolToSafeMode converts boolean to SafeMode enum (standardized across tests).
func BoolToSafeMode(b bool) enums.SafeMode {
	if b {
		return enums.SafeModeEnabled
	}

	return enums.SafeModeDisabled
}

// BoolToProfileStatus converts boolean to ProfileStatus enum (standardized across tests).
func BoolToProfileStatus(b bool) enums.ProfileStatus {
	if b {
		return enums.ProfileStatusEnabled
	}

	return enums.ProfileStatusDisabled
}

// BoolToOptimizationMode converts boolean to OptimizationMode enum (standardized across tests).
func BoolToOptimizationMode(b bool) enums.OptimizationMode {
	if b {
		return enums.OptimizationModeEnabled
	}

	return enums.OptimizationModeDisabled
}

// BoolToGenerationStatus converts boolean to GenerationStatus enum (standardized across tests).
func BoolToGenerationStatus(b bool) enums.GenerationStatus {
	if b {
		return enums.GenerationStatusCurrent
	}

	return enums.GenerationStatusHistorical
}

// ChainModifiers applies multiple config modifiers in sequence.
func ChainModifiers(modifiers ...ConfigModifier) ConfigModifier {
	return func(cfg *types.Config) *types.Config {
		for _, modifier := range modifiers {
			cfg = modifier(cfg)
		}

		return cfg
	}
}

// WithNixGenerationsSetting modifies a specific NixGenerations setting field.
// Takes the config, profile name, operation name, and a modifier function.
// Returns true if the operation was found and modified.
func WithNixGenerationsSetting(
	cfg *types.Config,
	profileName, operationName string,
	settingModifier func(*operations.NixGenerationsSettings) bool,
) bool {
	return WithOperationSettings(
		cfg,
		profileName,
		operationName,
		func(settings *operations.OperationSettings) bool {
			if settings.NixGenerations == nil {
				return false
			}

			return settingModifier(settings.NixGenerations)
		},
	)
}

// WithProfileOperationField modifies an operation field directly.
// Takes the config, profile name, operation name, and a modifier function.
// Returns true if the operation was found and modified.
func WithProfileOperationField(
	cfg *types.Config, profileName, operationName string,
	fieldModifier func(*types.CleanupOperation) bool,
) bool {
	return ModifyProfileOperation(
		cfg,
		profileName,
		operationName,
		func(profile *types.Profile, op *types.CleanupOperation) bool {
			return fieldModifier(op)
		},
	)
}
