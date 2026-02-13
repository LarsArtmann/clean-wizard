package domain

import (
	"testing"
)

// TestDefaultSettingsValidation verifies that DefaultSettings() validates enum values.
func TestDefaultSettingsValidation(t *testing.T) {
	t.Parallel()

	testCases := []OperationType{
		OperationTypeNixGenerations,
		OperationTypeTempFiles,
		OperationTypeHomebrew,
		OperationTypeNodePackages,
		OperationTypeGoPackages,
		OperationTypeCargoPackages,
		OperationTypeBuildCache,
		OperationTypeDocker,
		OperationTypeSystemCache,
		OperationTypeSystemTemp,
		OperationTypeProjectsManagementAutomation,
	}

	for _, opType := range testCases {
		t.Run(string(opType), func(t *testing.T) {
			// This should not panic for valid defaults
			settings := DefaultSettings(opType)
			if settings == nil {
				t.Fatalf("DefaultSettings(%s) returned nil", opType)
			}
		})
	}

	// Test that a custom type returns empty settings without panic
	customType := OperationType("custom-operation")
	settings := DefaultSettings(customType)
	if settings == nil {
		t.Fatal("DefaultSettings(custom) returned nil")
	}
}
