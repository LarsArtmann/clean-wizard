package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/execution"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCleanReportJSONExclusivity verifies --report and --json cannot combine
// on clean; the conflict fails fast (Rejection) before any cleaner runs.
func TestCleanReportJSONExclusivity(t *testing.T) {
	t.Parallel()

	err := runCleanCommand(nil, nil, true, false, true, true, "", "", "", 0, "", 0, false, "/tmp/should-never-be-written.html")

	require.Error(t, err)
	assert.Equal(t, errorfamily.Rejection, errorfamily.Classify(err))
	assert.Equal(t, "clean.report_json_conflict", errorfamily.Code(err))
	assert.NoFileExists(t, "/tmp/should-never-be-written.html")
}

// TestScanReportMachineOutputExclusivity verifies --report cannot combine with
// either machine-output format on scan.
func TestScanReportMachineOutputExclusivity(t *testing.T) {
	t.Parallel()

	for name, sarif := range map[string]bool{"json": false, "sarif": true} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := runScanCommand(false, "", !sarif, sarif, "", 0, "", 0, false, "/tmp/should-never-be-written.html")

			require.Error(t, err)
			assert.Equal(t, errorfamily.Rejection, errorfamily.Classify(err))
			assert.Equal(t, "scan.report_machine_output_conflict", errorfamily.Code(err))
		})
	}
}

// TestWriteReportFile_SelfContainedArtifact exercises the report write path
// end-to-end: fixture result → file on disk → valid, self-contained HTML.
func TestWriteReportFile_SelfContainedArtifact(t *testing.T) {
	t.Parallel()

	wr := &execution.WorkflowResult{
		Steps: []execution.StepResult{
			{
				Name: "report-target",
				Clean: types.CleanResult{
					SizeEstimate: types.SizeEstimate{Known: 2048, Status: enums.SizeEstimateStatusKnown},
					ItemsRemoved: 4,
				},
				Duration: 1500 * time.Millisecond,
			},
			{
				Name:     "report-skip",
				Err:      errorfamily.NewTransient("test.transient", "gone"),
				Duration: 10 * time.Millisecond,
			},
		},
		TotalBytesFreed: 2048,
		Duration:        2 * time.Second,
	}

	reportPath := filepath.Join(t.TempDir(), "report.html")
	require.NoError(t, writeReportFile(reportPath, wr, "clean run · dry-run: true", "clean"))

	raw, err := os.ReadFile(reportPath)
	require.NoError(t, err)

	html := string(raw)
	assert.True(t, strings.HasPrefix(html, "<!DOCTYPE html>"))
	assert.Contains(t, html, `"id":"report-target"`)
	assert.Contains(t, html, "clean run · dry-run: true")
	assert.Contains(t, html, "Content-Security-Policy")
	assert.NotContains(t, html, "<script src=", "self-contained: no external scripts")
}

// TestWriteReportFile_UnwritablePathIsRejection verifies error classification
// on the file-creation failure path.
func TestWriteReportFile_UnwritablePathIsRejection(t *testing.T) {
	t.Parallel()

	unwritable := filepath.Join(t.TempDir(), "missing-dir", "report.html")
	err := writeReportFile(unwritable, &execution.WorkflowResult{}, "subtitle", "clean")

	require.Error(t, err)
	assert.Equal(t, errorfamily.Rejection, errorfamily.Classify(err))
	assert.Equal(t, "clean.report_write", errorfamily.Code(err))
}
