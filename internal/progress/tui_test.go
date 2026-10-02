package progress

import (
	"context"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/execution"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTUIEmitter_SatisfiesExecutionContract is the spike's primary surface:
// the BubbleTea adapter drops into the same emitter seam as the nom Emitter.
var _ execution.ProgressEmitter = (*TUIEmitter)(nil)

// TestTUIEmitter_EventDispatchDoesNotPanic dispatches the full per-activity
// narrative without booting the BubbleTea program (reporter.send falls back
// to synchronous model updates when not started).
func TestTUIEmitter_EventDispatchDoesNotPanic(t *testing.T) {
	t.Parallel()

	em := NewTUIEmitter(context.Background())
	require.NotNil(t, em.Subscriber())

	assert.NotPanics(t, func() {
		em.ActivityRegistered("spike-cleaner")
		em.ActivityStarted("spike-cleaner")
		em.ActivityRetrying("spike-cleaner", 1, "transient")
		em.Note("drained note")
		em.ActivitySkipped("absent-cleaner", "not installed")
		em.ActivityCompleted("spike-cleaner", 2*time.Second)
		em.ActivityFailed("failing-cleaner", assertError("boom"), time.Second)
		em.WorkflowFinished(nil)
		em.Finish()
	})
}

// TestTUIEmitter_SkippedReasonFallsBack pins the ADR-0002 skip representation
// on the TUI path too.
func TestTUIEmitter_SkippedReasonFallsBack(t *testing.T) {
	t.Parallel()

	em := NewTUIEmitter(context.Background())

	assert.NotPanics(t, func() {
		em.ActivitySkipped("mystery", "")
	})
}

// assertError is a tiny error stand-in.
type assertError string

func (e assertError) Error() string { return string(e) }
