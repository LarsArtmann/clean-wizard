package execution

import (
	"context"
	"fmt"
	"time"

	flow "github.com/Azure/go-workflow"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/cockroachdb/errors"
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
}

// NewBuilder creates a Builder with the given options.
func NewBuilder(verbose bool) *Builder {
	return &Builder{verbose: verbose}
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
			return nil, errors.Newf("cleaner %q not found in registry", name)
		}

		collector.register(name, i)

		step := flow.FuncIO(
			name,
			makeCleanStepFunc(name, c, collector),
		)

		wf.Add(flow.Step(step).
			BeforeStep(before).
			AfterStep(after))
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
			return nil, errors.Newf("cleaner %q not found in registry", name)
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
				collector.record(name, domain.CleanResult{}, panicErr, duration)
				err = panicErr
				return
			}
		}()

		res := c.Clean(ctx)
		duration := time.Since(startTime)

		if res.IsErr() {
			collector.record(name, domain.CleanResult{}, res.Error(), duration)
			return domain.CleanResult{}, res.Error()
		}

		result = res.Value()
		collector.record(name, result, nil, duration)

		return result, nil
	}
}

// makeScanStepFunc creates a step function that wraps a cleaner's Scan method,
// recording the result in the collector as a CleanResult-equivalent.
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
				collector.record(name, domain.CleanResult{}, panicErr, duration)
				err = panicErr
				return
			}
		}()

		res := c.Scan(ctx)
		duration := time.Since(startTime)

		if res.IsErr() {
			collector.record(name, domain.CleanResult{}, res.Error(), duration)
			return nil, res.Error()
		}

		items = res.Value()
		var totalSize uint64
		for _, item := range items {
			totalSize += uint64(item.Size)
		}
		collector.record(name, domain.CleanResult{
			FreedBytes:   totalSize,
			ItemsRemoved: uint(len(items)),
		}, nil, duration)

		return items, nil
	}
}
