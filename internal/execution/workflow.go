package execution

import (
	"context"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/cockroachdb/errors"
)

// RunCleaners builds and executes a clean workflow for the given selected cleaners.
// It resolves cleaners from the registry, compiles them into a go-workflow DAG,
// executes it with the configured options, and returns aggregated results.
//
// The workflow runs steps in parallel up to maxConcurrency. Step errors are
// collected per-step (not short-circuited) so that one cleaner failure does
// not prevent others from running.
func RunCleaners(
	ctx context.Context,
	registry *cleaner.Registry,
	selected []string,
	opts ...RunOption,
) (*WorkflowResult, error) {
	cfg := resolveRunOptions(opts)

	builder := NewBuilder(cfg.verbose)
	if cfg.retry != nil {
		builder.WithRetryConfig(cfg.retry)
	}
	compiled, err := builder.BuildClean(registry, selected)
	if err != nil {
		return nil, err
	}

	return executeWorkflow(ctx, compiled, cfg)
}

// RunScans builds and executes a scan workflow for the given selected cleaners.
// Each cleaner's Scan method runs as a parallel workflow step.
func RunScans(
	ctx context.Context,
	registry *cleaner.Registry,
	selected []string,
	opts ...RunOption,
) (*WorkflowResult, error) {
	cfg := resolveRunOptions(opts)

	builder := NewBuilder(cfg.verbose)
	compiled, err := builder.BuildScan(registry, selected)
	if err != nil {
		return nil, err
	}

	return executeWorkflow(ctx, compiled, cfg)
}

// executeWorkflow configures and runs a compiled workflow, then aggregates results.
// Workflow-level errors (e.g. panics recovered by DontPanic) are attached to the
// WorkflowResult rather than silently dropped, so partial successes are preserved
// while still surfacing failures.
func executeWorkflow(ctx context.Context, compiled *CompiledWorkflow, cfg runConfig) (*WorkflowResult, error) {
	if cfg.maxConcurrency > 0 {
		compiled.Workflow.MaxConcurrency = cfg.maxConcurrency
	}

	startTime := time.Now()
	runErr := compiled.Workflow.Do(ctx)
	duration := time.Since(startTime)

	result := &WorkflowResult{
		Steps:    compiled.Collector.sortedByRegistration(),
		Duration: duration,
	}

	for _, step := range result.Steps {
		if step.Err == nil {
			result.TotalBytesFreed += step.Clean.FreedBytes
			result.TotalItemsRemoved += step.Clean.ItemsRemoved
		}
		result.TotalItemsFailed += step.Clean.ItemsFailed
	}

	if runErr != nil && len(result.Steps) == 0 {
		return nil, errors.Wrap(runErr, "workflow execution failed with no collected steps")
	}

	return result, nil
}
