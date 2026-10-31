package config

import (
	"errors"

	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	errorfamily "github.com/larsartmann/go-error-family"
)

// validateChangeBusinessRules validates changes against business rules.
// validateChangeBusinessRules applies the config-change business rules.
//
// NOTE (assessed 2026-10-31, do not rewrite): a go-business-rules conversion
// was evaluated and rejected — two rules with early-return semantics, an
// options-dependent confirmation gate, and a production-logging side effect
// gain nothing from the rules engine but indirection. Revisit only if
// change-level severities gain a real consumer (TODO 30).
func (vm *ValidationMiddleware) validateChangeBusinessRules(changes []ConfigChange) error {
	for _, change := range changes {
		// Rule: Cannot remove critical protected paths
		if change.Field == "protected" &&
			change.Operation == enums.ChangeOperationRemovedType {
			criticalPaths := types.CriticalSystemPaths()
			for _, critical := range criticalPaths {
				if change.OldValue == critical {
					return errorfamily.NewRejection(
						"config.protected_critical_removal",
						"cannot remove critical protected path: "+critical,
					)
				}
			}
		}

		// Rule: Cannot disable safe mode without explicit confirmation
		if change.Field == "safe_mode" && change.OldValue == true && change.NewValue == false { //nolint:goconst
			// Check if safe mode confirmation is required
			if vm.options != nil && vm.options.RequireSafeModeConfirmation {
				return errorfamily.NewRejection(
					"config.safe_mode_disable_unconfirmed",
					"disabling safe mode requires explicit confirmation (use --confirm-safe-mode-disable)",
				)
			}

			// Log warning in production environments
			if vm.options != nil && vm.options.Environment == "production" {
				vm.logger.LogError("safe_mode", "config_change",
					errors.New("safe mode disabled in production environment"))
			}
		}
	}

	return nil
}

// validateOperationSettings validates operation-specific settings with type safety.
func (vm *ValidationMiddleware) validateOperationSettings(
	operationName string,
	op types.CleanupOperation,
) error {
	// Use the already-validated settings from the operation
	if op.Settings == nil {
		return nil // Settings are optional
	}

	opType := operations.GetOperationType(operationName)

	return op.Settings.ValidateSettings(opType)
}
