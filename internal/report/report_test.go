package report_test

import (
	"encoding/json/v2"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/execution"
	"github.com/LarsArtmann/clean-wizard/internal/report"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-output/daghtml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixtureResult is a deterministic mixed-outcome workflow result.
func fixtureResult() *execution.WorkflowResult {
	return &execution.WorkflowResult{
		Steps: []execution.StepResult{
			{
				Name: "nix",
				Clean: types.CleanResult{
					SizeEstimate: types.SizeEstimate{Known: 1536, Status: enums.SizeEstimateStatusKnown},
					ItemsRemoved: 3,
				},
				Duration: 2 * time.Second,
			},
			{
				Name:     "homebrew",
				Err:      cleaner.NewNotAvailableError("homebrew", ""),
				Duration: time.Millisecond,
			},
			{
				Name:     "docker",
				Err:      errorfamily.NewTransient("test.transient", "docker daemon unreachable"),
				Duration: time.Second,
			},
		},
		TotalBytesFreed: 1536,
		Duration:        3 * time.Second,
	}
}

func TestWorkflowResultToDAG_StatusMapping(t *testing.T) {
	t.Parallel()

	dag := report.WorkflowResultToDAG(fixtureResult())

	require.Len(t, dag.Nodes, 3)
	assert.Empty(t, dag.Edges, "cleaners run in parallel: no edges")

	succeeded := dag.Nodes[0]
	assert.Equal(t, "nix", succeeded.ID)
	assert.Equal(t, "nix", succeeded.Label)
	assert.Equal(t, "var(--success)", succeeded.Color)
	assert.False(t, succeeded.Error)
	assert.Equal(t, "freed 1.5 KiB | 3 items | 2.0 s", succeeded.Tooltip)

	skipped := dag.Nodes[1]
	assert.Equal(t, "homebrew", skipped.ID)
	assert.Equal(t, "var(--info)", skipped.Color)
	assert.False(t, skipped.Error)
	assert.Equal(t, "skipped: homebrew not available", skipped.Tooltip)

	failed := dag.Nodes[2]
	assert.Equal(t, "docker", failed.ID)
	assert.Equal(t, "var(--error)", failed.Color)
	assert.True(t, failed.Error, "failed nodes carry the error dot")
	assert.Equal(t, "failed: [transient:test.transient] docker daemon unreachable", failed.Tooltip)
}

func TestWorkflowResultToDAG_EmptyResultIsEmptyDAG(t *testing.T) {
	t.Parallel()

	dag := report.WorkflowResultToDAG(&execution.WorkflowResult{})

	assert.True(t, dag.IsEmpty())
	assert.Equal(t, 0, dag.NodeCount())
}

// TestWorkflowResultToDAG_GoldenNodes pins the serialized node payload — the
// data consumers actually see inside the HTML's JSON block.
func TestWorkflowResultToDAG_GoldenNodes(t *testing.T) {
	t.Parallel()

	dag := report.WorkflowResultToDAG(fixtureResult())

	got, err := json.Marshal(dag.Nodes)
	require.NoError(t, err)

	goldenPath := filepath.Join("testdata", "report-nodes.golden.json")
	if *updateGolden {
		require.NoError(t, os.MkdirAll("testdata", 0o755))
		require.NoError(t, os.WriteFile(goldenPath, got, 0o600))
	}

	want, err := os.ReadFile(goldenPath)
	require.NoError(t, err, "golden file missing; run with -update-golden to create")

	assert.JSONEq(t, string(want), string(got))
}

func TestWorkflowSummary_CountsAllOutcomes(t *testing.T) {
	t.Parallel()

	assert.Equal(t,
		"3 cleaners | 1 succeeded | 1 skipped | 1 failed | freed 1.5 KiB in 3.0 s",
		report.WorkflowSummary(fixtureResult()),
	)
}

// TestWriteWorkflowHTML_SelfContainedDocument verifies the rendered artifact:
// a complete HTML document embedding the theme, the node data, and the
// interactive graph bootstrap — no external resources.
func TestWriteWorkflowHTML_SelfContainedDocument(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	require.NoError(t, report.WriteWorkflowHTML(&buf, fixtureResult(), "dry run · 2026-10-02"))

	html := buf.String()

	assert.True(t, strings.HasPrefix(html, "<!DOCTYPE html>"))
	assert.Contains(t, html, "<title>Clean Wizard</title>")
	assert.Contains(t, html, "dry run · 2026-10-02")
	assert.Contains(t, html, `"id":"nix"`, "node data embedded as JSON")
	assert.Contains(t, html, `"error":true`, "failed node carries the error flag")
	assert.Contains(t, html, "Content-Security-Policy", "self-contained: CSP meta present")
	assert.Contains(t, html, "initDAGGraph", "graph bootstrap script embedded")
	assert.Contains(t, html, "3 cleaners | 1 succeeded", "footer summary embedded")
	assert.NotContains(t, html, "<script src=", "no external scripts")
	assert.NotContains(t, html, "<link rel=", "no external stylesheets")
}

var updateGolden = flag.Bool("update-golden", false, "rewrite golden files")

// compile-time guard: the DAG type must remain the daghtml model.
var _ = daghtml.DAG{}
