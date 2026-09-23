package config

import (
	"fmt"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
)

// validateProtectedPathsConflict checks if operations might affect protected paths.
func (cv *ConfigValidator) validateProtectedPathsConflict(
	protected []string,
	op types.CleanupOperation,
) error {
	switch op.Name {
	case "temp-files":
		// Check if temp files cleanup might affect protected paths
		return cv.checkTempFilesConflict(protected, op)
	case "nix-generations":
		// Check if nix cleanup might affect protected paths
		return cv.checkNixConflict(protected, op)
	default:
		// Generic conflict check
		return nil // Skip unknown operations
	}
}

// checkTempFilesConflict checks for temp files conflicts.
func (cv *ConfigValidator) checkTempFilesConflict(
	protected []string,
	op types.CleanupOperation,
) error {
	if op.Settings != nil && op.Settings.TempFiles != nil {
		for _, exclude := range op.Settings.TempFiles.Excludes {
			for _, protectedPath := range protected {
				if strings.HasPrefix(exclude, protectedPath) ||
					strings.HasPrefix(protectedPath, exclude) {
					return fmt.Errorf(
						"temp files exclude '%s' conflicts with protected path '%s'",
						exclude,
						protectedPath,
					)
				}
			}
		}
	}

	return nil
}

// checkNixConflict checks for Nix conflicts.
func (cv *ConfigValidator) checkNixConflict(protected []string, _ types.CleanupOperation) error {
	// Nix operations typically affect /nix/store, check if protected paths overlap
	nixStorePath := "/nix/store"
	for _, protectedPath := range protected {
		if strings.HasPrefix(protectedPath, nixStorePath) ||
			strings.HasPrefix(nixStorePath, protectedPath) {
			return fmt.Errorf("nix operations may conflict with protected path '%s'", protectedPath)
		}
	}

	return nil
}

// findMaxRiskLevel finds the maximum risk level in configuration.
func (cv *ConfigValidator) findMaxRiskLevel(cfg *types.Config) enums.RiskLevelType {
	maxRisk := enums.RiskLevelLowType
	for _, profile := range cfg.Profiles {
		maxRisk = maxRiskLevelFromOperations(profile.Operations, maxRisk)
		if maxRisk == enums.RiskLevelCriticalType {
			return enums.RiskLevelCriticalType
		}
	}

	return maxRisk
}
