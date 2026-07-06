package commands

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRunCleanCommand_DryRun_JSON verifies the full clean command pipeline:
// flag parsing → config loading → DI container → registry → workflow → JSON output.
func TestRunCleanCommand_DryRun_JSON(t *testing.T) {
	err := runCleanCommand(
		nil,
		nil,
		true,  // dryRun
		false, // verbose
		true,  // jsonOutput
		true,  // skipConfirmation
		"",    // mode
		"",    // profile
		"",    // configPath
		0,     // retries
		0,     // concurrency
	)

	// Should not error — dry-run is safe and non-destructive
	require.NoError(t, err)
}
