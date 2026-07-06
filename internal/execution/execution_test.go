package execution

import (
	"context"
	osexec "os/exec"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCleaner is a test double for cleaner.Cleaner.
type mockCleaner struct {
	name     string
	avail    bool
	cleanRes result.Result[domain.CleanResult]
	scanRes  result.Result[[]domain.ScanItem]
}

func (m *mockCleaner) Name() string { return m.name }

func (m *mockCleaner) Type() domain.OperationType                                { return domain.OperationTypeCargoPackages }
func (m *mockCleaner) Clean(_ context.Context) result.Result[domain.CleanResult] { return m.cleanRes }
func (m *mockCleaner) IsAvailable(_ context.Context) bool                        { return m.avail }
func (m *mockCleaner) Scan(_ context.Context) result.Result[[]domain.ScanItem]   { return m.scanRes }

func TestRunCleaners_SuccessfulSteps(t *testing.T) {
	registry := cleaner.NewRegistry()

	successCleaner := &mockCleaner{
		name:     "success-cleaner",
		avail:    true,
		cleanRes: result.Ok(domain.CleanResult{FreedBytes: 1024, ItemsRemoved: 5}),
	}
	registry.Register("success-cleaner", successCleaner)

	wr, err := RunCleaners(context.Background(), registry, []string{"success-cleaner"})
	require.NoError(t, err)
	require.NotNil(t, wr)

	assert.Len(t, wr.Succeeded(), 1)
	assert.Empty(t, wr.Failed())
	assert.Empty(t, wr.Skipped())
	assert.Equal(t, uint64(1024), wr.TotalBytesFreed)
	assert.Equal(t, uint(5), wr.TotalItemsRemoved)
}

func TestRunCleaners_MixedResults(t *testing.T) {
	registry := cleaner.NewRegistry()

	registry.Register("success", &mockCleaner{
		name:     "success",
		avail:    true,
		cleanRes: result.Ok(domain.CleanResult{FreedBytes: 500, ItemsRemoved: 2}),
	})
	registry.Register("failed", &mockCleaner{
		name:     "failed",
		avail:    true,
		cleanRes: result.Err[domain.CleanResult](assertError("cleaner failed: disk error")),
	})
	registry.Register("skipped", &mockCleaner{
		name:     "skipped",
		avail:    true,
		cleanRes: result.Err[domain.CleanResult](&cleaner.NotAvailableError{CleanerName: "some-tool"}),
	})

	wr, err := RunCleaners(context.Background(), registry, []string{"success", "failed", "skipped"})
	require.NoError(t, err)
	require.NotNil(t, wr)

	assert.Len(t, wr.Succeeded(), 1)
	assert.Len(t, wr.Failed(), 1)
	assert.Len(t, wr.Skipped(), 1)
	assert.Equal(t, uint64(500), wr.TotalBytesFreed)
}

func TestRunCleaners_EmptySelection(t *testing.T) {
	registry := cleaner.NewRegistry()

	wr, err := RunCleaners(context.Background(), registry, []string{})
	require.NoError(t, err)
	require.NotNil(t, wr)
	assert.Empty(t, wr.Steps)
}

func TestRunCleaners_UnknownCleanerReturnsError(t *testing.T) {
	registry := cleaner.NewRegistry()

	_, err := RunCleaners(context.Background(), registry, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found in registry")
}

func TestRunScans_SuccessfulSteps(t *testing.T) {
	registry := cleaner.NewRegistry()

	registry.Register("scanner", &mockCleaner{
		name:  "scanner",
		avail: true,
		scanRes: result.Ok([]domain.ScanItem{
			{Size: 100},
			{Size: 200},
			{Size: 300},
		}),
	})

	wr, err := RunScans(context.Background(), registry, []string{"scanner"})
	require.NoError(t, err)
	require.NotNil(t, wr)

	assert.Len(t, wr.Succeeded(), 1)
	assert.Equal(t, uint64(600), wr.TotalBytesFreed)
	assert.Equal(t, uint(3), wr.TotalItemsRemoved)
}

func TestWorkflowResult_CleanResultsMap(t *testing.T) {
	wr := &WorkflowResult{
		Steps: []StepResult{
			{Name: "ok1", Clean: domain.CleanResult{FreedBytes: 100}, Err: nil},
			{Name: "ok2", Clean: domain.CleanResult{FreedBytes: 200}, Err: nil},
			{Name: "fail", Clean: domain.CleanResult{}, Err: assertError("some error")},
		},
	}

	m := wr.CleanResultsMap()
	assert.Len(t, m, 2)
	assert.Contains(t, m, "ok1")
	assert.Contains(t, m, "ok2")
	assert.NotContains(t, m, "fail")
}

func TestStepResult_StatusClassification(t *testing.T) {
	tests := []struct {
		name     string
		step     StepResult
		expected StepStatus
	}{
		{
			name:     "succeeded",
			step:     StepResult{Err: nil},
			expected: StepStatusSucceeded,
		},
		{
			name:     "failed with generic error",
			step:     StepResult{Err: assertError("disk corruption")},
			expected: StepStatusFailed,
		},
		{
			name:     "skipped with NotAvailableError",
			step:     StepResult{Err: &cleaner.NotAvailableError{CleanerName: "some-tool"}},
			expected: StepStatusSkipped,
		},
		{
			name:     "skipped with exec.ErrNotFound",
			step:     StepResult{Err: osexec.ErrNotFound},
			expected: StepStatusSkipped,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.step.Status())
		})
	}
}

func TestWithMaxConcurrency(t *testing.T) {
	cfg := resolveRunOptions([]RunOption{WithMaxConcurrency(5)})
	assert.Equal(t, 5, cfg.maxConcurrency)
}

func TestWithVerbose(t *testing.T) {
	cfg := resolveRunOptions([]RunOption{WithVerbose(true)})
	assert.True(t, cfg.verbose)

	cfg = resolveRunOptions([]RunOption{WithVerbose(false)})
	assert.False(t, cfg.verbose)
}

// assertError is a helper that wraps a string in an error for test assertions.
type testError string

func (e testError) Error() string { return string(e) }

func assertError(msg string) error { return testError(msg) }
