package execution

import (
	"context"
	"fmt"
	"time"

	flow "github.com/Azure/go-workflow"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/format"
)

// stepStartKey is used to store the step start time in the context via BeforeStep hooks.
type stepStartKey struct{}

// makeBeforeHook creates a BeforeStep hook that records the start time
// and optionally prints a debug message.
func makeBeforeHook(verbose bool) flow.BeforeStep {
	return func(ctx context.Context, step flow.Steper) (context.Context, error) {
		ctx = context.WithValue(ctx, stepStartKey{}, time.Now())
		if verbose {
			fmt.Printf("  [DEBUG] Running cleaner: %s\n", flow.String(step))
		}
		return ctx, nil
	}
}

// makeAfterHook creates an AfterStep hook that prints verbose output
// for successful steps. Error classification is handled by resultCollector.
func makeAfterHook(verbose bool) flow.AfterStep {
	return func(ctx context.Context, step flow.Steper, runErr error) error {
		name := flow.String(step)
		start, ok := ctx.Value(stepStartKey{}).(time.Time)
		if !ok {
			start = time.Now()
		}
		duration := time.Since(start)

		if verbose && runErr == nil {
			if fn, ok := step.(*flow.Function[struct{}, domain.CleanResult]); ok {
				r := fn.Output
				fmt.Printf(
					"  [DEBUG] %s: %d bytes (%s), %d items, took %s\n",
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
