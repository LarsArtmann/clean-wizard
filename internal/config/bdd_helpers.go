package config

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// BDDTestHelpers provides standardized utilities for BDD testing
// Eliminates duplicate patterns across test files

// ConfigModifier represents a function that modifies a config
type ConfigModifier func(*domain.Config) *domain.Config

// ProfileOperationModifier modifies specific profile operations
type ProfileOperationModifier func(*domain.Profile, *domain.CleanupOperation) bool // returns true if modified

// FindProfileOperation finds a specific operation within a profile by name
// Returns the operation index, or -1 if not found
func FindProfileOperation(cfg *domain.Config, profileName, operationName string) (*domain.Profile, int) {
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

// ModifyProfileOperation applies a modifier to a specific operation
// Returns true if the operation was found and modified
func ModifyProfileOperation(cfg *domain.Config, profileName, operationName string, modifier ProfileOperationModifier) bool {
	profile, opIndex := FindProfileOperation(cfg, profileName, operationName)
	if profile == nil || opIndex == -1 {
		return false
	}

	return modifier(profile, &profile.Operations[opIndex])
}

// WithOperationSettings applies a settings modifier to a specific operation
// Returns true if the operation was found and modified
func WithOperationSettings(cfg *domain.Config, profileName, operationName string, settingsModifier func(*domain.OperationSettings) bool) bool {
	return ModifyProfileOperation(cfg, profileName, operationName, func(profile *domain.Profile, op *domain.CleanupOperation) bool {
		if op.Settings == nil {
			return false
		}
		return settingsModifier(op.Settings)
	})
}

// BoolToSafeMode converts boolean to SafeMode enum (standardized across tests)
func BoolToSafeMode(b bool) domain.SafeMode {
	if b {
		return domain.SafeModeEnabled
	}
	return domain.SafeModeDisabled
}

// BoolToProfileStatus converts boolean to ProfileStatus enum (standardized across tests)
func BoolToProfileStatus(b bool) domain.ProfileStatus {
	if b {
		return domain.ProfileStatusEnabled
	}
	return domain.ProfileStatusDisabled
}

// BoolToOptimizationMode converts boolean to OptimizationMode enum (standardized across tests)
func BoolToOptimizationMode(b bool) domain.OptimizationMode {
	if b {
		return domain.OptimizationModeEnabled
	}
	return domain.OptimizationModeDisabled
}

// BoolToGenerationStatus converts boolean to GenerationStatus enum (standardized across tests)
func BoolToGenerationStatus(b bool) domain.GenerationStatus {
	if b {
		return domain.GenerationStatusCurrent
	}
	return domain.GenerationStatusHistorical
}

// ChainModifiers applies multiple config modifiers in sequence
func ChainModifiers(modifiers ...ConfigModifier) ConfigModifier {
	return func(cfg *domain.Config) *domain.Config {
		for _, modifier := range modifiers {
			cfg = modifier(cfg)
		}
		return cfg
	}
}
