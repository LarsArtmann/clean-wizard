package execution

import (
	"context"
	"fmt"
	"time"

	flow "github.com/Azure/go-workflow"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	errorfamily "github.com/larsartmann/go-error-family"
)

// CompiledWorkflow holds a ready-to-execute workflow and its result collector.
type CompiledWorkflow struct {
	Workflow  *flow.Workflow
	Collector *resultCollector
}

// Builder compiles a cleaner registry into a go-workflow DAG.
// It is DI-agnostic — it receives a *cleaner.Registry and selected names
// as plain parameters, matching BuildFlow's execution package design.
type Builder struct {
	verbose bool
	retry   *RetryConfig
}

// NewBuilder creates a Builder with the given options.
func NewBuilder(verbose bool) *Builder {
	return &Builder{verbose: verbose, retry: nil}
}

// WithRetryConfig enables per-step retry on the builder.
func (b *Builder) WithRetryConfig(cfg *RetryConfig) *Builder {
	b.retry = cfg
	return b
}

// BuildClean compiles a clean workflow from the given registry and selected cleaner names.
// Each selected cleaner becomes a parallel flow.FuncIO step with BeforeStep/AfterStep hooks.
func (b *Builder) BuildClean(registry *cleaner.Registry, selected []string) (*CompiledWorkflow, error) {
	collector := newResultCollector()
	wf := &flow.Workflow{
		DontPanic: true,
	}

	before := makeBeforeHook(b.verbose)
	after := makeAfterHook(b.verbose)

	for i, name := range selected {
		c, ok := registry.Get(name)
		if !ok {
			return nil, errorfamily.NewRejection(
				"cleaner.not_found",
				fmt.Sprintf("cleaner %q not found in registry", name),
			)
		}

		collector.register(name, i)

		step := flow.FuncIO(
			name,
			makeCleanStepFunc(name, c, collector),
		)

		stepBuilder := flow.Step(step).
			BeforeStep(before).
			AfterStep(after)

		if b.retry != nil {
			if opts := retryOptions(*b.retry); len(opts) > 0 {
				stepBuilder = stepBuilder.Retry(opts...)
			}
		}

		wf.Add(stepBuilder)
	}

	return &CompiledWorkflow{
		Workflow:  wf,
		Collector: collector,
	}, nil
}

// BuildScan compiles a scan workflow from the given registry and selected cleaner names.
// Each selected cleaner becomes a parallel flow.FuncIO step.
func (b *Builder) BuildScan(registry *cleaner.Registry, selected []string) (*CompiledWorkflow, error) {
	collector := newResultCollector()
	wf := &flow.Workflow{
		DontPanic: true,
	}

	for i, name := range selected {
		c, ok := registry.Get(name)
		if !ok {
			return nil, errorfamily.NewRejection(
				"cleaner.not_found",
				fmt.Sprintf("cleaner %q not found in registry", name),
			)
		}

		collector.register(name, i)

		step := flow.FuncIO(
			name,
			makeScanStepFunc(name, c, collector),
		)

		wf.Add(flow.Step(step))
	}

	return &CompiledWorkflow{
		Workflow:  wf,
		Collector: collector,
	}, nil
}

// makeCleanStepFunc creates a step function that wraps a cleaner's Clean method,
// recording the result in the collector. Panics are recovered and recorded as
// errors so that a panicking cleaner doesn't silently disappear from results.
//
// IMPORTANT: The collector.record() call is in the defer block, NOT in the
// function body. This ensures only the FINAL outcome is recorded — when
// go-workflow retries, only the last attempt's result is kept, preventing
// duplicate entries in the WorkflowResult.
func makeCleanStepFunc(
	name string,
	c cleaner.Cleaner,
	collector *resultCollector,
) func(context.Context, struct{}) (domain.CleanResult, error) {
	return func(ctx context.Context, _ struct{}) (result domain.CleanResult, err error) {
		startTime := time.Now()

		defer func() {
			duration := time.Since(startTime)
			if r := recover(); r != nil {
				panicErr := fmt.Errorf("cleaner %s panicked: %v", name, r)
				collector.recordFinal(name, domain.CleanResult{}, panicErr, duration)
				err = panicErr
				return
			}
			if err != nil {
				collector.recordFinal(name, domain.CleanResult{}, err, duration)
				return
			}
			collector.recordFinal(name, result, nil, duration)
		}()

		res := c.Clean(ctx)
		if res.IsErr() {
			return domain.CleanResult{}, res.Error()
		}

		result = res.Value()
		return result, nil
	}
}

// makeScanStepFunc creates a step function that wraps a cleaner's Scan method,
// recording the result in the collector as a CleanResult-equivalent.
// Uses recordFinal to prevent duplicate entries on retry.
func makeScanStepFunc(
	name string,
	c cleaner.Cleaner,
	collector *resultCollector,
) func(context.Context, struct{}) ([]domain.ScanItem, error) {
	return func(ctx context.Context, _ struct{}) (items []domain.ScanItem, err error) {
		startTime := time.Now()

		defer func() {
			duration := time.Since(startTime)
			if r := recover(); r != nil {
				panicErr := fmt.Errorf("scanner %s panicked: %v", name, r)
				collector.recordFinal(name, domain.CleanResult{}, panicErr, duration)
				err = panicErr
				return
			}
			if err != nil {
				collector.recordFinal(name, domain.CleanResult{}, err, duration)
				return
			}
			var totalSize uint64
			for _, item := range items {
				totalSize += uint64(item.Size)
			}
			collector.recordFinal(name, domain.CleanResult{
				FreedBytes:   totalSize,
				ItemsRemoved: uint(len(items)),
			}, nil, duration)
		}()

		res := c.Scan(ctx)
		if res.IsErr() {
			return nil, res.Error()
		}

		items = res.Value()
		return items, nil
	}
}
