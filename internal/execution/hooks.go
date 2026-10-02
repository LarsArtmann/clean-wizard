package execution

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	flow "github.com/Azure/go-workflow"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/format"
)

// stepStartKey is used to store the step start time in the context via BeforeStep hooks.
type stepStartKey struct{}

// verboseWriter is the sink for verbose output when no progress renderer is
// active. It is a variable so tests can capture verbose lines.
var verboseWriter io.Writer = os.Stdout

// verboseLine routes a verbose line through the progress renderer's log lane
// when one is active (so the live frame stays coherent), otherwise appends it
// to the plain writer. fmt.Print* is avoided to satisfy forbidigo and to keep
// every output site renderer-aware.
func verboseLine(em ProgressEmitter, format string, args ...any) {
	line := fmt.Sprintf(format, args...)
	if em != nil {
		em.Note(line)

		return
	}

	_, _ = fmt.Fprintln(verboseWriter, line)
}

// makeBeforeHook creates a BeforeStep hook that records the start time,
// emits the activity-started progress event, and optionally prints a debug
// message. The hook fires once per retry attempt, which re-marks the activity
// as running after a scheduled retry — exactly the narrative nom renders.
func makeBeforeHook(verbose bool, em ProgressEmitter) flow.BeforeStep {
	return func(ctx context.Context, step flow.Steper) (context.Context, error) {
		ctx = context.WithValue(ctx, stepStartKey{}, time.Now())

		name := flow.String(step)
		em.ActivityStarted(name)

		if verbose {
			verboseLine(em, "  [DEBUG] Running cleaner: %s", name)
		}

		return ctx, nil
	}
}

// makeAfterHook creates an AfterStep hook that prints verbose output
// for successful steps. Error classification is handled by resultCollector;
// terminal progress events are emitted once per step after the workflow
// finishes (see executeWorkflow) so retried attempts don't emit failures.
func makeAfterHook(verbose bool, em ProgressEmitter) flow.AfterStep {
	return func(ctx context.Context, step flow.Steper, runErr error) error {
		name := flow.String(step)

		start, ok := ctx.Value(stepStartKey{}).(time.Time)
		if !ok {
			start = time.Now()
		}

		duration := time.Since(start)

		if verbose && runErr == nil {
			if fn, ok := step.(*flow.Function[struct{}, types.CleanResult]); ok {
				r := fn.Output
				verboseLine(
					em,
					"  [DEBUG] %s: %d bytes (%s), %d items, took %s",
					name,
					r.FreedBytes,
					format.Bytes(int64(r.FreedBytes)),
					r.ItemsRemoved,
					format.Duration(duration),
				)
			}
		}

		return runErr
	}
}
