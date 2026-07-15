package execution

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunCleaners_RealRegistry_DryRun is an integration test that builds a real
// cleaner registry and runs a dry-run clean workflow through the go-workflow engine.
// It verifies the end-to-end path: DI registry → builder → workflow.Do → result aggregation.
func TestRunCleaners_RealRegistry_DryRun(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("integration test: uses real system cleaners (slow)")
	}
	registry, err := cleaner.DefaultRegistryWithConfig(false, true) // dryRun=true
	require.NoError(t, err)
	require.NotNil(t, registry)

	// Run with a single safe cleaner (cargo) that won't modify the system
	wr, err := RunCleaners(context.Background(), registry, []string{cleaner.CleanerCargo},
		WithVerbose(false),
	)
	require.NoError(t, err)
	require.NotNil(t, wr)

	// Cargo may or may not be available, but the workflow should complete
	// either way without error
	if len(wr.Steps) > 0 {
		assert.Equal(t, cleaner.CleanerCargo, wr.Steps[0].Name)
	}
}

// TestRunCleaners_PanicRecovery verifies that a panicking cleaner is recorded
// as a failed step rather than silently swallowed.
func TestRunCleaners_PanicRecovery(t *testing.T) {
	t.Parallel()
	registry := cleaner.NewRegistry()

	registry.Register("panic-cleaner", &mockCleaner{
		name:     "panic-cleaner",
		avail:    true,
		cleanRes: result.Result[domain.CleanResult]{},
	})

	// Override the clean function to panic
	panickingCleaner := &panicCleaner{name: "panic-cleaner"}
	registry.Register("panic-cleaner", panickingCleaner)

	wr, err := RunCleaners(context.Background(), registry, []string{"panic-cleaner"})
	require.NoError(t, err) // panic should be recovered, not returned as top-level error
	require.NotNil(t, wr)

	require.Len(t, wr.Steps, 1)
	assert.Contains(t, wr.Steps[0].Err.Error(), "panicked")
	assert.Equal(t, StepStatusFailed, wr.Steps[0].Status())
}

// TestRunCleaners_DeterministicOrdering verifies that results are returned in
// registration order even when steps complete out of order due to parallelism.
func TestRunCleaners_DeterministicOrdering(t *testing.T) {
	t.Parallel()
	registry := cleaner.NewRegistry()

	registry.Register("slow", &delayedMockCleaner{
		name:     "slow",
		avail:    true,
		cleanRes: result.Ok(domain.CleanResult{FreedBytes: 100}),
		delay:    50 * time.Millisecond,
	})
	registry.Register("fast", &delayedMockCleaner{
		name:     "fast",
		avail:    true,
		cleanRes: result.Ok(domain.CleanResult{FreedBytes: 200}),
		delay:    1 * time.Millisecond,
	})

	wr, err := RunCleaners(context.Background(), registry, []string{"slow", "fast"})
	require.NoError(t, err)
	require.Len(t, wr.Steps, 2)

	// "slow" was registered first, so it must be first in results
	// even though "fast" completes first
	assert.Equal(t, "slow", wr.Steps[0].Name)
	assert.Equal(t, "fast", wr.Steps[1].Name)
}

// TestRunCleaners_Retry verifies that retry options are wired correctly.
// A cleaner that fails twice then succeeds should ultimately show as succeeded.
func TestRunCleaners_Retry(t *testing.T) {
	t.Parallel()
	registry := cleaner.NewRegistry()

	failingThenSucceeding := &retryableMockCleaner{
		name:      "retry-me",
		avail:     true,
		failCount: 2,
	}
	registry.Register("retry-me", failingThenSucceeding)

	retryCfg := &RetryConfig{
		MaxAttempts:    3,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
	}

	wr, err := RunCleaners(context.Background(), registry, []string{"retry-me"},
		WithRetry(retryCfg),
	)
	require.NoError(t, err)
	require.NotNil(t, wr)

	// CRITICAL: despite 3 total attempts (2 failures + 1 success), there must be
	// exactly ONE step entry — recordFinal replaces, not appends.
	require.Len(t, wr.Steps, 1, "retried step must produce exactly 1 entry, not one per attempt")
	assert.Equal(t, "retry-me", wr.Steps[0].Name)
	assert.Equal(t, StepStatusSucceeded, wr.Steps[0].Status())
	assert.Equal(t, uint64(42), wr.Steps[0].Clean.FreedBytes)

	// Verify retries actually happened (2 failures before success)
	assert.Equal(t, int32(3), atomic.LoadInt32(&failingThenSucceeding.attempts))
}

// TestRunScans_RealRegistry_DryRun is an integration test for the scan workflow path.
func TestRunScans_RealRegistry_DryRun(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("integration test: uses real system cleaners (slow)")
	}
	registry, err := cleaner.DefaultRegistryWithConfig(false, true)
	require.NoError(t, err)

	wr, err := RunScans(context.Background(), registry, []string{cleaner.CleanerCargo})
	require.NoError(t, err)
	require.NotNil(t, wr)
}

// TestRunCleaners_SmartRetry_NotAvailable verifies that a cleaner returning
// NotAvailableError (classified as Infrastructure by go-error-family) is NOT
// retried — the NextBackOff hook returns backoff.Stop immediately. The cleaner
// must be called exactly once, not MaxAttempts times.
func TestRunCleaners_SmartRetry_NotAvailable(t *testing.T) {
	t.Parallel()
	registry := cleaner.NewRegistry()

	naCleaner := &countingMockCleaner{
		name:  "not-installed",
		avail: true,
		err:   cleaner.NewNotAvailableError("not-installed", ""),
	}
	registry.Register("not-installed", naCleaner)

	retryCfg := &RetryConfig{
		MaxAttempts:    5,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
	}

	start := time.Now()
	wr, err := RunCleaners(context.Background(), registry, []string{"not-installed"},
		WithRetry(retryCfg),
	)
	elapsed := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, wr)
	require.Len(t, wr.Steps, 1)

	// Infrastructure error → Skipped, not Failed
	assert.Equal(t, StepStatusSkipped, wr.Steps[0].Status())

	// Must be called exactly once — no retries for non-retryable errors
	assert.Equal(t, int32(1), atomic.LoadInt32(&naCleaner.attempts),
		"Infrastructure error must not be retried")

	// Should complete near-instantly (no backoff delay)
	assert.Less(t, elapsed, 100*time.Millisecond,
		"non-retryable error should not wait for backoff")
}

// TestRunCleaners_SmartRetry_Transient verifies that a cleaner returning a
// Transient error IS retried up to MaxAttempts. Uses errorfamily.NewTransient
// to classify the error as retryable.
func TestRunCleaners_SmartRetry_Transient(t *testing.T) {
	t.Parallel()
	registry := cleaner.NewRegistry()

	transientCleaner := &countingMockCleaner{
		name:      "transient-fail",
		avail:     true,
		err:       errorfamily.NewTransient("test.transient", "transient failure"),
		failCount: 3, // fail 3 times, then the retry budget is exhausted
	}
	registry.Register("transient-fail", transientCleaner)

	retryCfg := &RetryConfig{
		MaxAttempts:    3,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
	}

	wr, err := RunCleaners(context.Background(), registry, []string{"transient-fail"},
		WithRetry(retryCfg),
	)
	require.NoError(t, err)
	require.NotNil(t, wr)
	require.Len(t, wr.Steps, 1)

	// Transient error after retries exhausted → Failed (not Skipped)
	assert.Equal(t, StepStatusFailed, wr.Steps[0].Status())

	// Must be called exactly MaxAttempts times (3)
	assert.Equal(t, int32(3), atomic.LoadInt32(&transientCleaner.attempts),
		"Transient error must be retried up to MaxAttempts")
}

// --- Test Helpers ---

type panicCleaner struct {
	name string
}

func (p *panicCleaner) Name() string { return p.name }

func (p *panicCleaner) Type() domain.OperationType                                { return domain.OperationTypeCargoPackages }
func (p *panicCleaner) Clean(_ context.Context) result.Result[domain.CleanResult] { panic("boom") }
func (p *panicCleaner) IsAvailable(_ context.Context) bool                        { return true }
func (p *panicCleaner) Scan(_ context.Context) result.Result[[]domain.ScanItem] {
	return result.Ok([]domain.ScanItem{})
}

type delayedMockCleaner struct {
	name     string
	avail    bool
	cleanRes result.Result[domain.CleanResult]
	delay    time.Duration
}

func (d *delayedMockCleaner) Name() string               { return d.name }
func (d *delayedMockCleaner) Type() domain.OperationType { return domain.OperationTypeCargoPackages }
func (d *delayedMockCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	time.Sleep(d.delay)
	return d.cleanRes
}
func (d *delayedMockCleaner) IsAvailable(_ context.Context) bool { return d.avail }
func (d *delayedMockCleaner) Scan(_ context.Context) result.Result[[]domain.ScanItem] {
	return result.Ok([]domain.ScanItem{})
}

type retryableMockCleaner struct {
	name      string
	avail     bool
	failCount int32
	attempts  int32
}

func (r *retryableMockCleaner) Name() string               { return r.name }
func (r *retryableMockCleaner) Type() domain.OperationType { return domain.OperationTypeCargoPackages }
func (r *retryableMockCleaner) Clean(_ context.Context) result.Result[domain.CleanResult] {
	attempt := atomic.AddInt32(&r.attempts, 1)
	if attempt <= r.failCount {
		return result.Err[domain.CleanResult](fmt.Errorf("transient failure attempt %d", attempt))
	}
	return result.Ok(domain.CleanResult{FreedBytes: 42})
}
func (r *retryableMockCleaner) IsAvailable(_ context.Context) bool { return r.avail }
func (r *retryableMockCleaner) Scan(_ context.Context) result.Result[[]domain.ScanItem] {
	return result.Ok([]domain.ScanItem{})
}

// countingMockCleaner returns a fixed error until failCount is reached,
// then returns success. Tracks total call count for retry verification.
type countingMockCleaner struct {
	name      string
	avail     bool
	err       error // error to return (if non-nil, always fails with this)
	failCount int32 // number of times to fail before succeeding
	attempts  int32
}

func (c *countingMockCleaner) Name() string               { return c.name }
func (c *countingMockCleaner) Type() domain.OperationType { return domain.OperationTypeCargoPackages }
func (c *countingMockCleaner) Clean(_ context.Context) result.Result[domain.CleanResult] {
	attempt := atomic.AddInt32(&c.attempts, 1)
	if c.err != nil {
		// Always return the same error (for NotAvailable / Transient tests)
		return result.Err[domain.CleanResult](c.err)
	}
	if attempt <= c.failCount {
		return result.Err[domain.CleanResult](fmt.Errorf("failure attempt %d", attempt))
	}
	return result.Ok(domain.CleanResult{FreedBytes: 42})
}
func (c *countingMockCleaner) IsAvailable(_ context.Context) bool { return c.avail }
func (c *countingMockCleaner) Scan(_ context.Context) result.Result[[]domain.ScanItem] {
	return result.Ok([]domain.ScanItem{})
}
