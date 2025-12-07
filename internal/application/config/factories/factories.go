package factories

import (
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/domain/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// SetNixGenerationsCount sets the Nix generations count in the config
func SetNixGenerationsCount(cfg *config.Config, generations int) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	// Find nix-cleanup profile and its nix-generations operation
	profile, profileExists := cfg.Profiles["nix-cleanup"]
	if !profileExists {
		return fmt.Errorf("'nix-cleanup' profile not found in config")
	}

	operationFound := false
	for i, op := range profile.Operations {
		if op.Name == "nix-generations" {
			if op.Settings == nil {
				op.Settings = &shared.OperationSettings{}
			}
			if op.Settings.NixGenerations == nil {
				op.Settings.NixGenerations = &shared.NixGenerationsSettings{}
			}
			op.Settings.NixGenerations.Generations = generations
			profile.Operations[i] = op
			operationFound = true
			break
		}
	}

	if !operationFound {
		return fmt.Errorf("'nix-generations' operation not found in nix-cleanup profile")
	}

	return nil
}

// SetNixGenerationsOptimization sets the Nix generations optimization level in the config
func SetNixGenerationsOptimization(cfg *config.Config, optimizationLevel shared.OptimizationLevelType) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	// Find nix-cleanup profile and its nix-generations operation
	profile, profileExists := cfg.Profiles["nix-cleanup"]
	if !profileExists {
		return fmt.Errorf("'nix-cleanup' profile not found in config")
	}

	operationFound := false
	for i, op := range profile.Operations {
		if op.Name == "nix-generations" {
			if op.Settings == nil {
				op.Settings = &shared.OperationSettings{}
			}
			if op.Settings.NixGenerations == nil {
				op.Settings.NixGenerations = &shared.NixGenerationsSettings{}
			}
			op.Settings.NixGenerations.Generations = generations
			profile.Operations[i] = op
			operationFound = true
			break
		}
	}

	if !operationFound {
		return fmt.Errorf("'nix-generations' operation not found in nix-cleanup profile")
	}

	return nil
}
