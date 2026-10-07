package progress

import (
	"context"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/execution"
	"github.com/larsartmann/go-output/nom"
	"github.com/larsartmann/go-output/tui"
)

// TUIEmitter is the BubbleTea-backed execution.ProgressEmitter (ADR-0002 W3
// spike). It renders the same event narrative as Emitter but in an
// interactive BubbleTea program with zoom/scroll keybindings and ctrl+c
// cancellation of the underlying workflow context.
type TUIEmitter struct {
	// ctx carries the cancelable command context; cancel is registered with
	// the reporter so ctrl+c aborts the run gracefully.
	ctx    context.Context //nolint:containedctx // long-lived reporter, interface has no ctx
	report *tui.BubbleTeaProgressReporter
}

// compile-time proof the TUI spike satisfies the execution contract.
var _ execution.ProgressEmitter = (*TUIEmitter)(nil)

// NewTUIEmitter creates the BubbleTea reporter, registers cancel for graceful
// abort, and pins the NOM display mode (tree view, matching the nom frame).
func NewTUIEmitter(ctx context.Context) *TUIEmitter {
	ctx, cancel := context.WithCancel(ctx)
	reporter := tui.NewBubbleTeaProgressReporter()
	reporter.SetCancelFunc(cancel)
	reporter.SetDisplayMode(tui.DisplayModeNOM)

	return &TUIEmitter{ctx: ctx, report: reporter}
}

// Subscriber exposes the internal NOM subscriber for tests.
func (e *TUIEmitter) Subscriber() *nom.NOMSubscriber {
	return e.report.Subscriber()
}

// WorkflowStarted boots the BubbleTea program and announces the workflow.
func (e *TUIEmitter) WorkflowStarted(name string) {
	e.report.Start()
	_ = e.Subscriber().OnEvent(e.ctx, nom.WorkflowStarted{ID: nom.WorkflowID(name), Name: nom.WorkflowName(name)})
}

func (e *TUIEmitter) ActivityRegistered(name string) {
	_ = e.Subscriber().
		OnEvent(e.ctx, nom.ActivityRegistered{
			ID:       nom.ActivityID(name),
			Name:     nom.ActivityName(name),
			Kind:     nom.ActivityKindTask,
			Deps:     nil,
			Category: "",
		})
}

func (e *TUIEmitter) ActivityStarted(name string) {
	_ = e.Subscriber().
		OnEvent(e.ctx, nom.ActivityStarted{
			ID:       nom.ActivityID(name),
			Name:     nom.ActivityName(name),
			Kind:     nom.ActivityKindTask,
			Deps:     nil,
			Host:     "",
			Download: nom.DownloadProgress{},
			Category: "",
		})
}

func (e *TUIEmitter) ActivityRetrying(name string, attempt int, reason string) {
	_ = e.Subscriber().OnEvent(e.ctx, nom.ActivityRetrying{
		ID:      nom.ActivityID(name),
		Name:    nom.ActivityName(name),
		Attempt: attempt,
		Reason:  reason,
	})
}

func (e *TUIEmitter) ActivityCompleted(name string, duration time.Duration) {
	_ = e.Subscriber().OnEvent(e.ctx, nom.ActivityCompleted{
		ID:       nom.ActivityID(name),
		Name:     nom.ActivityName(name),
		Duration: duration,
	})
}

func (e *TUIEmitter) ActivityFailed(name string, err error, duration time.Duration) {
	_ = e.Subscriber().OnEvent(e.ctx, nom.ActivityFailed{
		ID:       nom.ActivityID(name),
		Name:     nom.ActivityName(name),
		Err:      err,
		Duration: duration,
	})
}

// ActivitySkipped follows the ADR-0002 representation: drained note + instant
// completion. The note renders in the TUI's message lane.
func (e *TUIEmitter) ActivitySkipped(name string, reason string) {
	if reason == "" {
		reason = "unavailable"
	}

	e.Note("• " + name + ": skipped (" + reason + ")")
	e.ActivityStarted(name)
	e.ActivityCompleted(name, time.Millisecond)
}

func (e *TUIEmitter) Note(line string) {
	e.report.ReportMessage(line)
}

// WorkflowFinished stops the reporter; the BubbleTea program quits in NOM mode.
func (e *TUIEmitter) WorkflowFinished(_ error) {
	e.report.Stop()
}

// Finish is a no-op: the TUI owns its final render and teardown via Stop.
func (e *TUIEmitter) Finish() {}
