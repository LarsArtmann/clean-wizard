// Package progress renders live workflow progress for interactive runs via
// go-output/nom (ADR-0002). The CLI layer constructs an Emitter only when the
// gating contract holds (TTY && !--json && !--sarif && --progress) and
// injects it into the execution layer via execution.WithProgress; every
// other mode stays byte-identical to a run without progress.
//
// Isolation: nom's timing cache is keyed by bare activity name and defaults
// to ~/.cache/nom-timing.csv — a path BuildFlow also uses. DefaultCachePath
// moves clean-wizard's cache to <user cache dir>/clean-wizard/nom-timing.csv
// so the two applications never poison each other's ETAs.
package progress

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/execution"
	"github.com/larsartmann/go-output/nom"
	"golang.org/x/term"
)

// refreshInterval is the renderer's frame rate; 100ms keeps the elapsed
// timers lively without burning CPU on redraws.
const refreshInterval = 100 * time.Millisecond

// cacheDirPerm is the permission for the timing-cache parent directory.
const cacheDirPerm = 0o755

// Emitter is the nom-backed execution.ProgressEmitter implementation.
// All methods are safe to call from concurrent workflow steps — nom's
// subscriber and renderer take their own locks.
type Emitter struct {
	// ctx is the command context the renderer is bound to for its entire
	// lifetime (struct field, not a parameter — the ProgressEmitter methods
	// have no ctx parameter).
	ctx      context.Context //nolint:containedctx // long-lived renderer, interface has no ctx
	sub      *nom.NOMSubscriber
	renderer *nom.InlineRenderer

	mu       sync.Mutex
	started  bool
	finished bool
}

// compile-time proof the nom-backed emitter satisfies the execution contract.
var _ execution.ProgressEmitter = (*Emitter)(nil)

// New creates an emitter rendering to out. Call Enabled first — a non-TTY
// writer would degrade to append-only plain text, which the gating contract
// (ADR-0002) reserves for explicit non-interactive choices.
func New(ctx context.Context, out io.Writer, appName string) *Emitter {
	sub := nom.NewNOMSubscriber(nom.WithCachePath(DefaultCachePath()))
	renderer := nom.NewInlineRenderer(sub, out, 0)
	renderer.SetAppName(appName)
	renderer.SetStartTime(time.Now())

	return &Emitter{
		ctx:      ctx,
		sub:      sub,
		renderer: renderer,
		mu:       sync.Mutex{},
		started:  false,
		finished: false,
	}
}

// Enabled reports whether live progress may activate: the writer must be a
// real terminal. CI environments are covered transitively (no *os.File TTY).
func Enabled(out io.Writer) bool {
	f, ok := out.(*os.File)
	if !ok {
		return false
	}

	return term.IsTerminal(int(f.Fd()))
}

// DefaultCachePath returns the isolated timing-cache location and ensures its
// parent directory exists. Falls back to the temp dir when the user cache dir
// is unavailable (matching nom's own fallback shape).
func DefaultCachePath() string {
	base, err := os.UserCacheDir()
	if err != nil || base == "" {
		base = os.TempDir()
	}

	dir := filepath.Join(base, "clean-wizard")
	if mkdirErr := os.MkdirAll(dir, cacheDirPerm); mkdirErr != nil {
		dir = filepath.Join(os.TempDir(), "clean-wizard")
		_ = os.MkdirAll(dir, cacheDirPerm)
	}

	return filepath.Join(dir, "nom-timing.csv")
}

// WorkflowStarted announces the workflow and starts the refresh loop.
func (e *Emitter) WorkflowStarted(name string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	_ = e.sub.OnEvent(e.ctx, nom.WorkflowStarted{ID: nom.WorkflowID(name), Name: nom.WorkflowName(name)})

	if !e.started {
		e.renderer.Start(e.ctx, refreshInterval)
		e.started = true
	}
}

// ActivityRegistered pre-creates an activity as pending so the full plan is
// visible before work begins.
func (e *Emitter) ActivityRegistered(name string) {
	_ = e.sub.OnEvent(e.ctx, nom.ActivityRegistered{ //nolint:exhaustruct_v5 // optional nom fields are absent by design
		ID:       nom.ActivityID(name),
		Name:     nom.ActivityName(name),
		Kind:     nom.ActivityKindTask,
		Deps:     nil,
		Category: "",
	})
}

// ActivityStarted marks an activity running; fires per retry attempt, which
// re-marks the activity as running after a scheduled retry.
func (e *Emitter) ActivityStarted(name string) {
	_ = e.sub.OnEvent(e.ctx, nom.ActivityStarted{ //nolint:exhaustruct_v5 // optional nom fields are absent by design
		ID:       nom.ActivityID(name),
		Name:     nom.ActivityName(name),
		Kind:     nom.ActivityKindTask,
		Deps:     nil,
		Host:     "",
		Download: nom.DownloadProgress{},
		Category: "",
	})
}

// ActivityRetrying renders the ⟳N (reason) suffix. Intermediate attempts are
// never reported as failures — nom would record that duration into its
// timing cache and skew ETAs (ADR-0002).
func (e *Emitter) ActivityRetrying(name string, attempt int, reason string) {
	_ = e.sub.OnEvent(e.ctx, nom.ActivityRetrying{
		ID:      nom.ActivityID(name),
		Name:    nom.ActivityName(name),
		Attempt: attempt,
		Reason:  reason,
	})
}

// ActivityCompleted finalizes an activity with its observed duration.
func (e *Emitter) ActivityCompleted(name string, duration time.Duration) {
	_ = e.sub.OnEvent(e.ctx, nom.ActivityCompleted{
		ID:       nom.ActivityID(name),
		Name:     nom.ActivityName(name),
		Duration: duration,
	})
}

// ActivityFailed finalizes an activity that exhausted its retries.
func (e *Emitter) ActivityFailed(name string, err error, duration time.Duration) {
	_ = e.sub.OnEvent(e.ctx, nom.ActivityFailed{
		ID:       nom.ActivityID(name),
		Name:     nom.ActivityName(name),
		Err:      err,
		Duration: duration,
	})
}

// ActivitySkipped reports an unavailable cleaner per the ADR-0002 decision:
// a drained explanatory note above the frame plus an instant completion —
// nom has no skip status, and a row that silently vanishes reads as a bug.
func (e *Emitter) ActivitySkipped(name string, reason string) {
	if reason == "" {
		reason = "unavailable"
	}

	e.Note("• " + name + ": skipped (" + reason + ")")
	e.ActivityStarted(name)
	e.ActivityCompleted(name, time.Millisecond)
}

// Note drains a log line above the live frame (nom log interleaving).
func (e *Emitter) Note(line string) {
	e.renderer.EnqueueLines([]string{line})
}

// WorkflowFinished stops the refresh loop; a nil error means success. The
// final tree stays on screen (with any failure annotation) until Finish.
func (e *Emitter) WorkflowFinished(_ error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.renderer.Stop()
}

// Finish prints the final static tree and restores terminal state. It is
// idempotent; executeWorkflow calls it exactly once per run, and the CLI may
// call it defensively on early error paths.
func (e *Emitter) Finish() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.finished {
		return
	}

	e.finished = true
	e.renderer.Finish()
}
