package commands

import (
	"encoding/json/v2"
	"io"
	"os"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/execution"
	"github.com/stretchr/testify/require"
)

func TestBuildScanResultsRecordsStepErrors(t *testing.T) {
	t.Parallel()

	notAvailable := cleaner.NewNotAvailableError("cargo", "")
	workflowResult := &execution.WorkflowResult{
		Steps: []execution.StepResult{
			{Name: "go", Clean: domain.CleanResult{ItemsRemoved: 2}, Err: nil, Duration: time.Second},
			{Name: "cargo", Err: notAvailable, Duration: time.Second},
		},
	}

	available := []CleanerConfig{
		{Name: "Go", Type: CleanerTypeGoPackages},
		{Name: "Cargo", Type: CleanerTypeCargoPackages},
	}

	results := buildScanResults(workflowResult, available)
	require.Len(t, results, 2)

	byName := make(map[string]ScanResult, len(results))
	for _, r := range results {
		byName[r.Name] = r
	}

	require.ErrorIs(t, byName["Cargo"].Err, notAvailable)
	require.NoError(t, byName["Go"].Err)
}

// captureStdout redirects os.Stdout while fn runs and returns what was printed.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout

	r, stdoutW, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = stdoutW

	defer func() { os.Stdout = old }()

	fn()
	require.NoError(t, stdoutW.Close())

	out, err := io.ReadAll(r)
	require.NoError(t, err)

	return string(out)
}

// TestOutputScanJSONEnrichesErrors must stay serial: captureStdout swaps
// os.Stdout, which would race with other parallel tests printing to stdout.
func TestOutputScanJSONEnrichesErrors(t *testing.T) {
	results := []ScanResult{
		{
			Name:           "Go",
			Available:      CleanerAvailabilityAvailable,
			ItemsCount:     3,
			BytesCleanable: 1024,
			Err:            nil,
		},
		{
			Name:      "Cargo",
			Available: CleanerAvailabilityAvailable,
			Err:       cleaner.NewNotAvailableError("cargo", ""),
		},
	}

	var out string

	require.NotPanics(t, func() {
		out = captureStdout(t, func() {
			require.NoError(t, outputScanJSON(results, 1024, 3))
		})
	})

	var parsed struct {
		Results []struct {
			Name      string `json:"name"`
			Items     uint   `json:"items"`
			Bytes     uint64 `json:"bytes"`
			Available bool   `json:"available"`
			Error     string `json:"error"`
			Family    string `json:"family"`
			Code      string `json:"code"`
			Retryable bool   `json:"retryable"`
		} `json:"results"`
		Summary struct {
			TotalBytes uint64 `json:"totalBytes"`
			TotalItems uint   `json:"totalItems"`
		} `json:"summary"`
	}

	require.NoError(t, json.Unmarshal([]byte(out), &parsed))
	require.Len(t, parsed.Results, 2)
	require.Equal(t, uint64(1024), parsed.Summary.TotalBytes)

	successful := parsed.Results[0]
	require.Equal(t, "Go", successful.Name)
	require.Empty(t, successful.Error, "successful result must not carry error fields")
	require.Empty(t, successful.Family)
	require.False(t, successful.Retryable)

	errored := parsed.Results[1]
	require.Equal(t, "Cargo", errored.Name)
	require.NotEmpty(t, errored.Error)
	require.Equal(t, "infrastructure", errored.Family)
	require.Equal(t, "cleaner.cargo.not_available", errored.Code)
	require.False(t, errored.Retryable, "Infrastructure errors are not retryable")
}
