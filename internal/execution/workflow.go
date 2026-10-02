package execution

import (
	"context"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	errorfamily "github.com/larsartmann/go-error-family"
)

// newRunBuilder resolves run options and configures a builder from them,
// applying retry settings when configured. Shared by RunCleaners and RunScans.
func newRunBuilder(opts []RunOption) (*Builder, runConfig) {
	cfg := resolveRunOptions(opts)

	builder := NewBuilder(cfg.verbose)
	builder.WithProgressEmitter(cfg.emitter())
	if cfg.retry != nil {
		builder.WithRetryConfig(cfg.retry)
	}

	return builder, cfg
}

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
	builder, cfg := newRunBuilder(opts)

	compiled, err := builder.BuildClean(registry, selected)
	if err != nil {
		return nil, err
	}

	return executeWorkflow(ctx, compiled, cfg, "clean")
}

// RunScans builds and executes a scan workflow for the given selected cleaners.
// Each cleaner's Scan method runs as a parallel workflow step.
func RunScans(
	ctx context.Context,
	registry *cleaner.Registry,
	selected []string,
	opts ...RunOption,
) (*WorkflowResult, error) {
	builder, cfg := newRunBuilder(opts)

	compiled, err := builder.BuildScan(registry, selected)
	if err != nil {
		return nil, err
	}

	return executeWorkflow(ctx, compiled, cfg, "scan")
}

// executeWorkflow configures and runs a compiled workflow, then aggregates results.
// Workflow-level errors (e.g. panics recovered by DontPanic) are attached to the
// WorkflowResult rather than silently dropped, so partial successes are preserved
// while still surfacing failures.
//
// Progress emission (ADR-0002): WorkflowStarted and the upfront ActivityRegistered
// batch announce the plan before work begins; per-attempt starts and retries are
// emitted from the step hooks and the retry scheduler; terminal per-step outcomes
// are derived once from the final collector state after the workflow finishes —
// the single place that knows whether retries are exhausted.
func executeWorkflow(ctx context.Context, compiled *CompiledWorkflow, cfg runConfig, workflowName string) (*WorkflowResult, error) {
	em := cfg.emitter()
	em.WorkflowStarted(workflowName)
	for _, name := range compiled.Collector.registeredNames() {
		em.ActivityRegistered(name)
	}

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

		emitTerminalOutcome(em, step)
	}

	if runErr != nil {
		em.WorkflowFailed(runErr)
	} else {
		em.WorkflowCompleted()
	}

	em.Finish()

	if runErr != nil && len(result.Steps) == 0 {
		return nil, errorfamily.WrapTransient(
			runErr,
			"execution.workflow_failed",
			"workflow execution failed with no collected steps",
		)
	}

	return result, nil
}

// emitTerminalOutcome reports a step's FINAL outcome exactly once, after the
// workflow has finished. Intermediate retry attempts are never reported as
// failures (ADR-0002): a transient failure shows as ActivityRetrying from the
// retry scheduler, and only the exhausted budget produces ActivityFailed.
func emitTerminalOutcome(em ProgressEmitter, step StepResult) {
	switch step.Status() {
	case StepStatusSucceeded:
		em.ActivityCompleted(step.Name, step.Duration)
	case StepStatusSkipped:
		em.ActivitySkipped(step.Name, reasonForSkip(step.Err))
	case StepStatusFailed:
		em.ActivityFailed(step.Name, step.Err, step.Duration)
	}
}

// reasonForSkip extracts a short human reason from an infrastructure-family
// error; the empty reason falls back to a stable, non-blaming default.
func reasonForSkip(err error) string {
	if err != nil && err.Error() != "" {
		return err.Error()
	}

	return "unavailable"
}
