package execution

import "time"

// ProgressEmitter receives workflow and activity lifecycle notifications for
// a live progress renderer. The interface deliberately mirrors the go-output
// nom event vocabulary WITHOUT importing it: internal/execution stays
// renderer-agnostic (ADR-0002), the nom-backed implementation lives in
// internal/progress, and the CLI layer injects it via WithProgress.
//
// Emission contract (ADR-0002 binding rules):
//   - ActivityFailed is emitted ONLY for final outcomes (retries exhausted),
//     never for intermediate attempts — nom records every ActivityFailed
//     duration into its timing cache, which would pollute ETAs.
//   - ActivityRetrying is emitted directly when a retry is scheduled; nom
//     resets the activity to running with a retry-count suffix.
//   - Unavailable cleaners are reported via ActivitySkipped: nom has no skip
//     status, so the implementation completes the activity instantly and
//     drains an explanatory note above the live frame.
type ProgressEmitter interface {
	WorkflowStarted(name string)
	ActivityRegistered(name string)
	ActivityStarted(name string)
	ActivityRetrying(name string, attempt int, reason string)
	ActivityCompleted(name string, duration time.Duration)
	ActivityFailed(name string, err error, duration time.Duration)
	ActivitySkipped(name string, reason string)
	Note(line string)
	// WorkflowFinished terminates the run-level narrative; a nil error means
	// success. Terminal per-step outcomes are emitted separately (once, from
	// the final collector state).
	WorkflowFinished(err error)
	Finish()
}

// noopProgressEmitter is the silent default used whenever no renderer was
// injected. All execution call sites emit through it, so behavior is
// byte-identical to the pre-progress implementation in every non-progress
// mode (ADR-0002 safety rail #1).
type noopProgressEmitter struct{}

func (noopProgressEmitter) WorkflowStarted(string)                      {}
func (noopProgressEmitter) ActivityRegistered(string)                   {}
func (noopProgressEmitter) ActivityStarted(string)                      {}
func (noopProgressEmitter) ActivityRetrying(string, int, string)        {}
func (noopProgressEmitter) ActivityCompleted(string, time.Duration)     {}
func (noopProgressEmitter) ActivityFailed(string, error, time.Duration) {}
func (noopProgressEmitter) ActivitySkipped(string, string)              {}
func (noopProgressEmitter) Note(string)                                 {}
func (noopProgressEmitter) WorkflowFinished(error)                      {}
func (noopProgressEmitter) Finish()                                     {}

// The single shared no-op instance; stateless, so sharing is safe.
var noopEmitter ProgressEmitter = noopProgressEmitter{}
