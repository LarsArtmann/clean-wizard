package commands

import (
	"testing"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOutputScanSARIFProducesEnvelope must stay serial: captureStdout swaps
// os.Stdout, which would race with other parallel tests printing to stdout.
func TestOutputScanSARIFProducesEnvelope(t *testing.T) {
	results := []ScanResult{
		{
			Name:           "Nix",
			RegistryName:   "nix",
			Available:      CleanerAvailabilityAvailable,
			ItemsCount:     4,
			BytesCleanable: 512,
		},
	}

	var out string

	require.NotPanics(t, func() {
		out = captureStdout(t, func() {
			require.NoError(t, outputScanSARIF(results))
		})
	})

	assert.Contains(t, out, `"version": "2.1.0"`)
	assert.Contains(t, out, `"ruleId": "nix"`)
	assert.Contains(t, out, `cleaner://nix`)
	assert.Contains(t, out, `"name": "clean-wizard"`)
}

func TestRunScanCommandRejectsConflictingOutputFlags(t *testing.T) {
	err := runScanCommand(false, "", true, true, "", 3, "", 0)
	require.Error(t, err)

	assert.Equal(t, errorfamily.Rejection, errorfamily.Classify(err))
	assert.Equal(t, "scan.flags", errorfamily.Code(err))
	assert.Contains(t, err.Error(), "--json and --sarif are mutually exclusive")
}
