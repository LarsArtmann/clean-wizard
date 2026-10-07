package execution

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/result"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// emitterEventKind enumerates the ProgressEmitter callbacks a recording
// emitter captures, so tests can assert full event sequences.
type emitterEventKind int

const (
	evWorkflowStarted emitterEventKind = iota
	evActivityRegistered
	evActivityStarted
	evActivityRetrying
	evActivityCompleted
	evActivityFailed
	evActivitySkipped
	evNote
	evWorkflowFinished
	evFinish
)

// emitterEvent is one captured emitter callback with its payload.
type emitterEvent struct {
	kind     emitterEventKind
	name     string
	attempt  int
	reason   string
	err      error
	duration time.Duration
}

func (e emitterEvent) String() string {
	switch e.kind {
	case evWorkflowStarted:
		return "WorkflowStarted(" + e.name + ")"
	case evActivityRegistered:
		return "ActivityRegistered(" + e.name + ")"
	case evActivityStarted:
		return "ActivityStarted(" + e.name + ")"
	case evActivityRetrying:
		return fmt.Sprintf("ActivityRetrying(%s, attempt=%d, reason=%s)", e.name, e.attempt, e.reason)
	case evActivityCompleted:
		return fmt.Sprintf("ActivityCompleted(%s, %s)", e.name, e.duration)
	case evActivityFailed:
		return "ActivityFailed(" + e.name + ", err=" + e.err.Error() + ")"
	case evActivitySkipped:
		return "ActivitySkipped(" + e.name + ", reason=" + e.reason + ")"
	case evNote:
		return "Note(" + e.reason + ")"
	case evWorkflowFinished:
		if e.err != nil {
			return "WorkflowFinished(" + e.err.Error() + ")"
		}

		return "WorkflowFinished(nil)"
	case evFinish:
		return "Finish"
	default:
		return fmt.Sprintf("Unknown(%d)", e.kind)
	}
}

// recordingEmitter is a thread-safe ProgressEmitter test double capturing the
// full event stream for sequence assertions. Safe under -race because
// go-workflow steps emit concurrently.
type recordingEmitter struct {
	mu     sync.Mutex
	events []emitterEvent
}

func newRecordingEmitter() *recordingEmitter {
	return &recordingEmitter{events: []emitterEvent{}}
}

func (r *recordingEmitter) record(e emitterEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.events = append(r.events, e)
}

func (r *recordingEmitter) WorkflowStarted(name string) {
	r.record(emitterEvent{kind: evWorkflowStarted, name: name})
}

func (r *recordingEmitter) ActivityRegistered(name string) {
	r.record(emitterEvent{kind: evActivityRegistered, name: name})
}

func (r *recordingEmitter) ActivityStarted(name string) {
	r.record(emitterEvent{kind: evActivityStarted, name: name})
}

func (r *recordingEmitter) ActivityRetrying(name string, attempt int, reason string) {
	r.record(emitterEvent{kind: evActivityRetrying, name: name, attempt: attempt, reason: reason})
}

func (r *recordingEmitter) ActivityCompleted(name string, duration time.Duration) {
	r.record(emitterEvent{kind: evActivityCompleted, name: name, duration: duration})
}

func (r *recordingEmitter) ActivityFailed(name string, err error, duration time.Duration) {
	r.record(emitterEvent{kind: evActivityFailed, name: name, err: err, duration: duration})
}

func (r *recordingEmitter) ActivitySkipped(name string, reason string) {
	r.record(emitterEvent{kind: evActivitySkipped, name: name, reason: reason})
}

func (r *recordingEmitter) Note(line string) {
	r.record(emitterEvent{kind: evNote, reason: line})
}

func (r *recordingEmitter) WorkflowFinished(err error) {
	r.record(emitterEvent{kind: evWorkflowFinished, err: err})
}

func (r *recordingEmitter) Finish() {
	r.record(emitterEvent{kind: evFinish})
}

// snapshot copies the captured event slice under the lock.
func (r *recordingEmitter) snapshot() []emitterEvent {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]emitterEvent, len(r.events))
	copy(out, r.events)

	return out
}

// filter returns only the events matching the given kinds, in emission order.
func (r *recordingEmitter) filter(kinds ...emitterEventKind) []emitterEvent {
	wanted := make(map[emitterEventKind]bool, len(kinds))
	for _, k := range kinds {
		wanted[k] = true
	}

	var out []emitterEvent

	for _, e := range r.snapshot() {
		if wanted[e.kind] {
			out = append(out, e)
		}
	}

	return out
}

// count returns how many events of the given kind were emitted.
func (r *recordingEmitter) count(kind emitterEventKind) int {
	return len(r.filter(kind))
}

// countFor returns how many events of the given kind name the given activity.
func (r *recordingEmitter) countFor(kind emitterEventKind, name string) int {
	n := 0

	for _, e := range r.filter(kind) {
		if e.name == name {
			n++
		}
	}

	return n
}

// kinds returns the kind sequence of all captured events as strings, for
// readable assertion failures.
func (r *recordingEmitter) dump() string {
	var b strings.Builder

	for i, e := range r.snapshot() {
		if i > 0 {
			b.WriteString("\n")
		}

		b.WriteString(e.String())
	}

	return b.String()
}

// TestRunCleaners_Progress_HappyPathSequence asserts the exact event narrative
// for two successful cleaners running sequentially: workflow boundary, the
// registered-upfront batch, per-step started, then terminal completions emitted
// once from the final collector state.
func TestRunCleaners_Progress_HappyPathSequence(t *testing.T) {
	t.Parallel()

	registry := cleaner.NewRegistry()
	registry.Register("alpha", &mockCleaner{
		name:     "alpha",
		avail:    true,
		cleanRes: result.Ok(types.CleanResult{FreedBytes: 10, ItemsRemoved: 1}),
	})
	registry.Register("beta", &mockCleaner{
		name:     "beta",
		avail:    true,
		cleanRes: result.Ok(types.CleanResult{FreedBytes: 20, ItemsRemoved: 2}),
	})

	em := newRecordingEmitter()
	wr, err := RunCleaners(context.Background(), registry, []string{"alpha", "beta"},
		WithProgress(em),
		WithMaxConcurrency(1),
	)
	require.NoError(t, err)
	require.NotNil(t, wr)

	events := em.snapshot()
	expected := []emitterEvent{
		{kind: evWorkflowStarted, name: "clean"},
		{kind: evActivityRegistered, name: "alpha"},
		{kind: evActivityRegistered, name: "beta"},
		{kind: evActivityStarted, name: "alpha"},
		{kind: evActivityStarted, name: "beta"},
		{kind: evActivityCompleted, name: "alpha"},
		{kind: evActivityCompleted, name: "beta"},
		{kind: evWorkflowFinished, err: nil},
		{kind: evFinish},
	}
	require.Len(t, events, len(expected), "event stream mismatch:\n%s", em.dump())

	for i, want := range expected {
		assert.Equal(t, want.kind, events[i].kind, "event %d kind mismatch:\n%s", i, em.dump())
		assert.Equal(t, want.name, events[i].name, "event %d name mismatch:\n%s", i, em.dump())
	}

	// Registration is announced exactly once per cleaner — a duplicate would
	// re-render the plan row in the live tree.
	for _, name := range []string{"alpha", "beta"} {
		assert.Equal(t, 1, em.countFor(evActivityRegistered, name), "exactly one registration for %s", name)
	}

	// Terminal completions carry the collector's observed duration.
	for _, e := range em.filter(evActivityCompleted) {
		assert.GreaterOrEqual(t, e.duration, time.Duration(0))
	}
}

// TestRunCleaners_Progress_NoEmitterIsSilent verifies the nil-safety contract:
// WithProgress(nil) and omitting the option entirely both run to completion
// without any progress wiring.
func TestRunCleaners_Progress_NoEmitterIsSilent(t *testing.T) {
	t.Parallel()

	registry := cleaner.NewRegistry()
	registry.Register("plain", &mockCleaner{
		name:     "plain",
		avail:    true,
		cleanRes: result.Ok(types.CleanResult{FreedBytes: 5}),
	})

	for _, opts := range [][]RunOption{
		nil,
		{WithProgress(nil)},
	} {
		wr, err := RunCleaners(context.Background(), registry, []string{"plain"}, opts...)
		require.NoError(t, err)
		require.NotNil(t, wr)
		require.Len(t, wr.Succeeded(), 1)
	}

	// The unresolved default must be a usable non-nil emitter so call sites
	// never nil-check.
	assert.NotNil(t, resolveRunOptions(nil).emitter())
}

// TestRunCleaners_Progress_RetryEmitsRetryingNotFailed asserts the retry
// narrative: every scheduled retry emits ActivityRetrying with the 1-based
// attempt number and the error-family reason, each attempt re-emits
// ActivityStarted, and NO ActivityFailed is emitted for intermediate attempts
// (ADR-0002: nom would pollute its timing cache with failed-attempt durations).
func TestRunCleaners_Progress_RetryEmitsRetryingNotFailed(t *testing.T) {
	t.Parallel()

	registry := cleaner.NewRegistry()

	flaky := &retryableMockCleaner{name: "flaky", avail: true, failCount: 2}
	registry.Register("flaky", flaky)

	em := newRecordingEmitter()
	wr, err := RunCleaners(context.Background(), registry, []string{"flaky"},
		WithProgress(em),
		WithRetry(&RetryConfig{MaxAttempts: 3, InitialBackoff: time.Millisecond, MaxBackoff: 10 * time.Millisecond}),
		WithMaxConcurrency(1),
	)
	require.NoError(t, err)
	require.NotNil(t, wr)
	assert.Equal(t, StepStatusSucceeded, wr.Steps[0].Status())

	started := em.filter(evActivityStarted)
	retrying := em.filter(evActivityRetrying)
	failed := em.filter(evActivityFailed)

	require.Len(t, started, 3, "one Started per attempt:\n%s", em.dump())
	require.Len(t, retrying, 2, "one Retrying per scheduled retry:\n%s", em.dump())
	assert.Empty(t, failed, "intermediate attempts must never emit ActivityFailed")

	assert.Equal(t, 1, retrying[0].attempt)
	assert.Equal(t, 2, retrying[1].attempt)
	assert.Equal(t, "transient", retrying[0].reason, "reason is the errorfamily family name")

	// Narrative order: Started → Retrying(1) → Started → Retrying(2) → Started → Completed.
	sequence := em.filter(evActivityStarted, evActivityRetrying, evActivityCompleted)
	require.Len(t, sequence, 6)

	wantKinds := []emitterEventKind{
		evActivityStarted, evActivityRetrying,
		evActivityStarted, evActivityRetrying,
		evActivityStarted, evActivityCompleted,
	}
	for i, want := range wantKinds {
		assert.Equal(t, want, sequence[i].kind, "sequence position %d:\n%s", i, em.dump())
	}
}

// TestRunCleaners_Progress_ExhaustedRetriesEmitsSingleFailure asserts that a
// permanently-transient cleaner produces exactly ONE ActivityFailed — emitted
// after the retry budget is exhausted, from the final collector state.
func TestRunCleaners_Progress_ExhaustedRetriesEmitsSingleFailure(t *testing.T) {
	t.Parallel()

	registry := cleaner.NewRegistry()

	alwaysTransient := &countingMockCleaner{
		name:  "always-transient",
		avail: true,
		err:   errorfamily.NewTransient("test.transient", "transient failure"),
	}
	registry.Register("always-transient", alwaysTransient)

	em := newRecordingEmitter()
	wr, err := RunCleaners(context.Background(), registry, []string{"always-transient"},
		WithProgress(em),
		WithRetry(&RetryConfig{MaxAttempts: 3, InitialBackoff: time.Millisecond, MaxBackoff: 10 * time.Millisecond}),
		WithMaxConcurrency(1),
	)
	require.NoError(t, err)
	require.NotNil(t, wr)
	assert.Equal(t, StepStatusFailed, wr.Steps[0].Status())

	assert.Equal(t, 3, em.count(evActivityStarted), "one Started per attempt")
	assert.Equal(t, 2, em.count(evActivityRetrying), "NextBackOff fires only when another attempt WILL run")
	assert.Equal(t, 0, em.count(evActivityCompleted))

	failed := em.filter(evActivityFailed)
	require.Len(t, failed, 1, "exactly one terminal failure:\n%s", em.dump())
	require.Error(t, failed[0].err)

	// The failure is terminal: it must come after every retry event.
	lastRetrying := -1
	failedAt := -1

	for i, e := range em.snapshot() {
		switch e.kind {
		case evActivityRetrying:
			lastRetrying = i
		case evActivityFailed:
			failedAt = i
		case evActivityCompleted:
			t.Fatalf("unexpected ActivityCompleted:\n%s", em.dump())
		}
	}

	assert.Greater(t, failedAt, lastRetrying, "terminal failure must follow all retries")
}

// TestRunCleaners_Progress_SkipEmitsSkippedWithReason asserts the ADR-0002 skip
// representation: an unavailable cleaner emits one ActivitySkipped carrying the
// error message as reason, and neither Completed nor Failed.
func TestRunCleaners_Progress_SkipEmitsSkippedWithReason(t *testing.T) {
	t.Parallel()

	registry := cleaner.NewRegistry()

	unavailable := &countingMockCleaner{
		name:  "unavailable",
		avail: true,
		err:   cleaner.NewNotAvailableError("some-tool", ""),
	}
	registry.Register("unavailable", unavailable)

	em := newRecordingEmitter()
	wr, err := RunCleaners(context.Background(), registry, []string{"unavailable"},
		WithProgress(em),
		WithRetry(&RetryConfig{MaxAttempts: 3, InitialBackoff: time.Millisecond, MaxBackoff: 10 * time.Millisecond}),
		WithMaxConcurrency(1),
	)
	require.NoError(t, err)
	require.NotNil(t, wr)
	assert.Equal(t, StepStatusSkipped, wr.Steps[0].Status())
	assert.Equal(t, int32(1), unavailable.attempts.Load(), "Infrastructure errors are not retried")

	assert.Equal(t, 1, em.count(evActivityStarted))
	assert.Equal(t, 0, em.count(evActivityCompleted))
	assert.Equal(t, 0, em.count(evActivityFailed))
	assert.Equal(t, 0, em.count(evActivityRetrying))

	skipped := em.filter(evActivitySkipped)
	require.Len(t, skipped, 1)
	assert.Equal(t, "unavailable", skipped[0].name)
	assert.Equal(t, "some-tool not available", skipped[0].reason)
}

// TestRunScans_Progress_EmitsWorkflowNarrative verifies the scan path shares
// the same emission contract (the scan builder wires hooks through the same
// executeWorkflow boundary).
func TestRunScans_Progress_EmitsWorkflowNarrative(t *testing.T) {
	t.Parallel()

	registry := cleaner.NewRegistry()
	registry.Register("scanner", &mockCleaner{
		name:    "scanner",
		avail:   true,
		scanRes: result.Ok([]types.ScanItem{}),
	})

	em := newRecordingEmitter()
	wr, err := RunScans(context.Background(), registry, []string{"scanner"}, WithProgress(em))
	require.NoError(t, err)
	require.NotNil(t, wr)

	assert.Equal(t, 1, em.count(evWorkflowStarted))
	assert.Equal(t, "scan", em.filter(evWorkflowStarted)[0].name)
	assert.Equal(t, 1, em.count(evActivityRegistered))
	assert.Equal(t, 1, em.count(evActivityStarted))
	assert.Equal(t, 1, em.count(evActivityCompleted))
	assert.Equal(t, 1, em.count(evFinish))
	assert.NoError(t, em.filter(evWorkflowFinished)[0].err)
}

// TestRunCleaners_Progress_ParallelEmittersAreRaceSafe runs several cleaners
// concurrently against one emitter; under -race any unsynchronized emitter
// access fails. Order across cleaners is nondeterministic, so only boundary
// invariants are asserted.
func TestRunCleaners_Progress_ParallelEmittersAreRaceSafe(t *testing.T) {
	t.Parallel()

	registry := cleaner.NewRegistry()

	names := []string{"p1", "p2", "p3", "p4"}
	for _, name := range names {
		registry.Register(name, &mockCleaner{
			name:     name,
			avail:    true,
			cleanRes: result.Ok(types.CleanResult{FreedBytes: 1}),
		})
	}

	em := newRecordingEmitter()
	wr, err := RunCleaners(context.Background(), registry, names, WithProgress(em))
	require.NoError(t, err)
	require.NotNil(t, wr)

	events := em.snapshot()
	require.NotEmpty(t, events)
	assert.Equal(t, evWorkflowStarted, events[0].kind, "workflow boundary must be first")
	assert.Equal(t, evFinish, events[len(events)-1].kind, "Finish must be last")

	assert.Equal(t, len(names), em.count(evActivityRegistered))
	assert.Equal(t, len(names), em.count(evActivityStarted))
	assert.Equal(t, len(names), em.count(evActivityCompleted))
	assert.Equal(t, 1, em.count(evWorkflowFinished))
	assert.NoError(t, em.filter(evWorkflowFinished)[0].err)

	registered := em.filter(evActivityRegistered)
	for i, name := range names {
		assert.Equal(t, name, registered[i].name, "registration must follow selection order")
	}
}

// TestEmitTerminalOutcome_StatusMapping is table-driven over the
// step-status-to-event mapping: succeeded → Completed, Infrastructure →
// Skipped, every other family → Failed.
func TestEmitTerminalOutcome_StatusMapping(t *testing.T) {
	t.Parallel()

	notAvailable := cleaner.NewNotAvailableError("some-tool", "")
	transient := errorfamily.NewTransient("test.transient", "boom")
	rejection := errorfamily.NewRejection("test.rejection", "bad input")

	tests := []struct {
		name     string
		step     StepResult
		wantKind emitterEventKind
		wantName string
		wantDur  time.Duration
	}{
		{
			name:     "succeeded step emits ActivityCompleted",
			step:     StepResult{Name: "ok", Duration: 2 * time.Second},
			wantKind: evActivityCompleted,
			wantName: "ok",
			wantDur:  2 * time.Second,
		},
		{
			name:     "infrastructure error emits ActivitySkipped",
			step:     StepResult{Name: "skip", Err: notAvailable, Duration: time.Millisecond},
			wantKind: evActivitySkipped,
			wantName: "skip",
			wantDur:  0, // ActivitySkipped carries a reason, not a duration
		},
		{
			name:     "transient error emits ActivityFailed",
			step:     StepResult{Name: "fail", Err: transient, Duration: time.Second},
			wantKind: evActivityFailed,
			wantName: "fail",
			wantDur:  time.Second,
		},
		{
			name:     "rejection error emits ActivityFailed",
			step:     StepResult{Name: "reject", Err: rejection, Duration: time.Second},
			wantKind: evActivityFailed,
			wantName: "reject",
			wantDur:  time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			em := newRecordingEmitter()
			emitTerminalOutcome(em, tt.step)

			matched := em.filter(tt.wantKind)
			require.Len(t, matched, 1)
			assert.Equal(t, tt.wantName, matched[0].name)
			assert.Equal(t, tt.wantDur, matched[0].duration)
		})
	}
}

// TestReasonForSkip covers the skip-reason fallback: a real message is passed
// through; a nil or empty-message error falls back to a stable default.
func TestReasonForSkip(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "some-tool not available", reasonForSkip(cleaner.NewNotAvailableError("some-tool", "")))
	assert.Equal(t, "unavailable", reasonForSkip(nil))
	assert.Equal(t, "unavailable", reasonForSkip(errors.New("")))
}

// TestVerboseLineRouting verifies the renderer-aware verbose output: with an
// emitter, verbose lines drain through Note; without one, they append to the
// verbose writer (never fmt.Print* — forbidigo).
func TestVerboseLineRouting(t *testing.T) {
	t.Run("emitter active routes to Note", func(t *testing.T) {
		t.Parallel()

		em := newRecordingEmitter()
		verboseLine(em, "count %d", 7)

		notes := em.filter(evNote)
		require.Len(t, notes, 1)
		assert.Equal(t, "count 7", notes[0].reason)
	})

	t.Run("no emitter appends to writer", func(t *testing.T) {
		var buf syncBuffer

		original := verboseWriter
		verboseWriter = &buf

		t.Cleanup(func() { verboseWriter = original })

		verboseLine(nil, "value %s", "x")
		assert.Equal(t, "value x\n", buf.String())
	})
}

// syncBuffer is a minimal mutex-guarded io.Writer for writer-capture tests.
type syncBuffer struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}
