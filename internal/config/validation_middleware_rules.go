package config

import (
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// validateChangeBusinessRules validates changes against business rules
func (vm *ValidationMiddleware) validateChangeBusinessRules(changes []ConfigChange) error {
	for _, change := range changes {
		// Rule: Cannot remove critical protected paths
		if change.Field == "protected" && change.Operation == OperationRemoved {
			criticalPaths := []string{"/", "/System", "/usr", "/etc"}
			for _, critical := range criticalPaths {
				if change.OldValue == critical {
					return fmt.Errorf("cannot remove critical protected path: %s", critical)
				}
			}
		}

		// Rule: Cannot disable safe mode without explicit confirmation
		if change.Field == "safe_mode" && change.OldValue == true && change.NewValue == false {
			// Check if safe mode confirmation is required
			if vm.options != nil && vm.options.RequireSafeModeConfirmation {
				return fmt.Errorf("disabling safe mode requires explicit confirmation (use --confirm-safe-mode-disable)")
			}
			
			// Log warning in production environments
			if vm.options != nil && vm.options.Environment == "production" {
				vm.logger.LogError("safe_mode", "config_change", 
					fmt.Errorf("safe mode disabled in production environment"))
			}
		}
	}

	return nil
}

// validateOperationSettings validates operation-specific settings with type safety
func (vm *ValidationMiddleware) validateOperationSettings(operationName string, op domain.CleanupOperation) error {
	// Use the already-validated settings from the operation
	if op.Settings == nil {
		return nil // Settings are optional
	}

	opType := domain.GetOperationType(operationName)
	return op.Settings.ValidateSettings(opType)
}
