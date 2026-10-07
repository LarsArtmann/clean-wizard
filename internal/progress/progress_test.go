package progress

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/execution"
	"github.com/LarsArtmann/clean-wizard/internal/result"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// silentCleaner is a trivial cleaner double for workflow-driven tests.
type silentCleaner struct {
	name  string
	delay time.Duration
}

func (s *silentCleaner) Name() string { return s.name }
func (s *silentCleaner) Type() operations.OperationType {
	return operations.OperationTypeCargoPackages
}

func (s *silentCleaner) Clean(ctx context.Context) result.Result[types.CleanResult] {
	select {
	case <-time.After(s.delay):
	case <-ctx.Done():
	}

	return result.Ok(types.CleanResult{
		SizeEstimate: types.SizeEstimate{Known: 64, Status: enums.SizeEstimateStatusKnown},
		ItemsRemoved: 2,
	})
}

func (s *silentCleaner) IsAvailable(_ context.Context) bool { return true }

func (s *silentCleaner) Scan(_ context.Context) result.Result[[]types.ScanItem] {
	return result.Ok([]types.ScanItem{})
}

// lockedBuffer is a concurrency-safe bytes.Buffer (the nom renderer draws from
// its own refresh goroutine).
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}

func (b *lockedBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Len()
}

// captureStdout redirects os.Stdout into a pipe for the duration of run and
// returns everything written to it.
func captureStdout(t *testing.T, run func()) string {
	t.Helper()

	original := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = writer

	var wg sync.WaitGroup
	wg.Add(1)

	var captured []byte

	go func() {
		defer wg.Done()

		captured, _ = io.ReadAll(reader)
	}()

	run()

	writer.Close()

	os.Stdout = original

	wg.Wait()
	reader.Close()

	return string(captured)
}

// TestWorkflowRun_ProgressNeverWritesToStdout is the M45 guarantee: a run with
// an ACTIVE nom-backed emitter writes zero bytes to the process stdout — the
// renderer owns its own sink — so --json/--sarif streams stay byte-clean even
// if a future wiring bug ever constructed an emitter in machine mode.
func TestWorkflowRun_ProgressNeverWritesToStdout(t *testing.T) {
	// No t.Parallel: swapping the os.Stdout global races with any concurrent
	// test doing the same.
	sink := &lockedBuffer{}
	em := New(context.Background(), sink, "clean-wizard-test")

	registry := cleaner.NewRegistry()
	registry.Register("sink-test-cleaner", &silentCleaner{name: "sink-test-cleaner", delay: 150 * time.Millisecond})

	var runErr error

	stdout := captureStdout(t, func() {
		_, runErr = execution.RunCleaners(
			context.Background(),
			registry,
			[]string{"sink-test-cleaner"},
			execution.WithProgress(em),
		)
	})

	require.NoError(t, runErr)
	assert.Empty(t, stdout, "progress rendering must never touch process stdout")

	// The renderer was genuinely active: its own sink received frames naming
	// the activity (guards against a vacuously-green zero-byte assertion).
	assert.Contains(t, sink.String(), "sink-test-cleaner")
}

// TestWorkflowRun_NoEmitterWritesNothingToStdout is the flip side: without a
// progress emitter the execution layer is fully silent on stdout.
func TestWorkflowRun_NoEmitterWritesNothingToStdout(t *testing.T) {
	// No t.Parallel: swaps the os.Stdout global.
	registry := cleaner.NewRegistry()
	registry.Register("quiet-cleaner", &silentCleaner{name: "quiet-cleaner"})

	var runErr error

	stdout := captureStdout(t, func() {
		_, runErr = execution.RunCleaners(context.Background(), registry, []string{"quiet-cleaner"})
	})

	require.NoError(t, runErr)
	assert.Empty(t, stdout)
}

// TestEnabled_NonTTYWritersAreRejected pins the CI half of the gate: anything
// that is not a terminal *os.File is disabled.
func TestEnabled_NonTTYWritersAreRejected(t *testing.T) {
	t.Parallel()

	assert.False(t, Enabled(&bytes.Buffer{}), "bytes.Buffer is not a TTY")
	assert.False(t, Enabled(io.Discard), "io.Discard is not a TTY")
	assert.False(t, Enabled(nil), "nil is not a TTY")
}

// TestDefaultCachePath_IsolatedUnderAppDir verifies the timing-cache isolation
// rule (ADR-0002): the path lives under a clean-wizard directory, never at
// nom's shared default (~/.cache/nom-timing.csv which BuildFlow also uses),
// and the parent directory is created.
func TestDefaultCachePath_IsolatedUnderAppDir(t *testing.T) {
	t.Parallel()

	path := DefaultCachePath()

	assert.Contains(t, path, "clean-wizard", "cache must be namespaced under the app dir")
	assert.Contains(t, path, "nom-timing.csv")
	assert.DirExists(t, filepath.Dir(path), "parent directory must exist")
}
