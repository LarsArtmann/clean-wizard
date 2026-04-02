// Package result provides functional programming patterns for error handling and flow control.
//
// FlowBuilder and Pipeline types enable building complex pipelines with sequential execution,
// branching, parallel execution, and comprehensive error handling.
//
// Key types:
//   - FlowBuilder[T]: Builder for constructing executable pipelines
//   - Pipeline[T]: Composed pipeline ready for execution
//   - ParallelFlow[T]: Concurrent executor for parallel operations
//
// Example usage:
//
//	pipeline := NewFlowBuilder[CleanResult]().
//	    Step("scan", func(ctx context.Context) Result[CleanResult] { return Scan(ctx) }).
//	    Step("validate", func(ctx context.Context, r CleanResult) Result[CleanResult] { return Validate(ctx, r) })
//
//	result := pipeline.Execute(ctx)
package result

import (
	"context"
	"sync"
)

// Step represents a single step in a pipeline.
type Step[T any] struct {
	Name    string
	Execute func(context.Context) Result[T]
}

// FlowBuilder enables building complex pipelines with branching, parallel execution, and error handling.
type FlowBuilder[T any] struct {
	steps      []Step[T]
	maxRetries int
	timeout    any // context.WithTimeout duration
}

// NewFlowBuilder creates a new FlowBuilder for the specified result type.
func NewFlowBuilder[T any]() *FlowBuilder[T] {
	return &FlowBuilder[T]{
		steps:      []Step[T]{},
		maxRetries: 0,
		timeout:    nil,
	}
}

// Step adds a named step to the pipeline.
func (fb *FlowBuilder[T]) Step(name string, fn func(context.Context) Result[T]) *FlowBuilder[T] {
	fb.steps = append(fb.steps, Step[T]{
		Name:    name,
		Execute: fn,
	})

	return fb
}

// StepWithRetry adds a step with automatic retry on failure.
func (fb *FlowBuilder[T]) StepWithRetry(
	name string,
	maxRetries int,
	fn func(context.Context) Result[T],
) *FlowBuilder[T] {
	fb.maxRetries = maxRetries

	return fb.Step(name, func(ctx context.Context) Result[T] {
		var lastErr error

		for attempt := 0; attempt <= maxRetries; attempt++ {
			res := fn(ctx)
			if res.IsOk() {
				return res
			}

			lastErr = res.Error()
		}

		return Err[T](lastErr)
	})
}

// Then chains a new step that receives the output of the previous step.
func (fb *FlowBuilder[T]) Then(
	name string,
	fn func(context.Context, T) Result[T],
) *FlowBuilder[T] {
	// Capture the index of the previous step BEFORE adding the new step
	prevIdx := len(fb.steps) - 1

	return fb.Step(name, func(ctx context.Context) Result[T] {
		if prevIdx < 0 {
			return Err[T](ErrNoPreviousStep)
		}

		lastStep := fb.steps[prevIdx]
		prevResult := lastStep.Execute(ctx)

		if prevResult.IsErr() {
			return prevResult
		}

		return fn(ctx, prevResult.Value())
	})
}

// ParallelFlow is a concurrent executor.
type ParallelFlow[T any] struct {
	steps    []Step[T]
	mu       sync.RWMutex
	results  sync.Map
	executed bool
}

// NewParallelFlow creates a new ParallelFlow for concurrent execution.
func NewParallelFlow[T any]() *ParallelFlow[T] {
	return &ParallelFlow[T]{
		steps: []Step[T]{},
	}
}

// Add adds a step to the parallel flow.
func (pf *ParallelFlow[T]) Add(name string, fn func(context.Context) Result[T]) *ParallelFlow[T] {
	pf.steps = append(pf.steps, Step[T]{
		Name:    name,
		Execute: fn,
	})

	return pf
}

// Execute runs all steps concurrently and returns a map of results.
func (pf *ParallelFlow[T]) Execute(ctx context.Context) map[string]Result[T] {
	if len(pf.steps) == 0 {
		pf.mu.Lock()
		pf.executed = true
		pf.mu.Unlock()

		return make(map[string]Result[T])
	}

	// Create a WaitGroup for synchronization
	var wg sync.WaitGroup
	wg.Add(len(pf.steps))

	// Execute each step in a goroutine
	for _, step := range pf.steps {
		go func(s Step[T]) {
			defer wg.Done()

			pf.results.Store(s.Name, s.Execute(ctx))
		}(step)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Convert sync.Map to regular map
	results := make(map[string]Result[T], len(pf.steps))
	pf.results.Range(func(key, value any) bool {
		results[key.(string)] = value.(Result[T])

		return true
	})

	pf.mu.Lock()
	pf.executed = true
	// Store the results map for later access
	for k, v := range results {
		pf.results.Store(k, v)
	}
	pf.mu.Unlock()

	return results
}

// Results returns the results after execution.
func (pf *ParallelFlow[T]) Results() map[string]Result[T] {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	if !pf.executed {
		return make(map[string]Result[T])
	}

	results := make(map[string]Result[T])
	pf.results.Range(func(key, value any) bool {
		results[key.(string)] = value.(Result[T])

		return true
	})

	return results
}

// Successful returns only the successful results.
func (pf *ParallelFlow[T]) Successful() map[string]T {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	successful := make(map[string]T)

	pf.results.Range(func(key, value any) bool {
		result := value.(Result[T])
		if result.IsOk() {
			successful[key.(string)] = result.Value()
		}

		return true
	})

	return successful
}

// Failed returns only the failed results.
func (pf *ParallelFlow[T]) Failed() map[string]error {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	failed := make(map[string]error)

	pf.results.Range(func(key, value any) bool {
		result := value.(Result[T])
		if result.IsErr() {
			failed[key.(string)] = result.Error()
		}

		return true
	})

	return failed
}

// ErrNoPreviousStep is returned when a Then() is called on an empty FlowBuilder.
var ErrNoPreviousStep = &NoPreviousStepError{}

// NoPreviousStepError indicates that Then() was called on an empty FlowBuilder.
type NoPreviousStepError struct{}

func (e *NoPreviousStepError) Error() string {
	return "cannot chain step: no previous step in pipeline"
}

// Pipeline represents a composed pipeline that can be executed.
type Pipeline[T any] struct {
	steps    []Step[T]
	metadata map[string]string
}

// Execute runs all steps in sequence, returning the final result.
func (p *Pipeline[T]) Execute(ctx context.Context) Result[T] {
	if len(p.steps) == 0 {
		return Err[T](ErrEmptyPipeline)
	}

	var lastResult Result[T]

	for _, step := range p.steps {
		lastResult = step.Execute(ctx)
		if lastResult.IsErr() {
			return lastResult
		}
	}

	return lastResult
}

// ErrEmptyPipeline is returned when executing an empty pipeline.
var ErrEmptyPipeline = &EmptyPipelineError{}

// EmptyPipelineError indicates that the pipeline has no steps.
type EmptyPipelineError struct{}

func (e *EmptyPipelineError) Error() string {
	return "pipeline has no steps"
}

// Build converts the FlowBuilder to an executable Pipeline.
func (fb *FlowBuilder[T]) Build() *Pipeline[T] {
	return &Pipeline[T]{
		steps:    fb.steps,
		metadata: make(map[string]string),
	}
}

// Execute runs the pipeline and returns the final result.
func (fb *FlowBuilder[T]) Execute(ctx context.Context) Result[T] {
	return fb.Build().Execute(ctx)
}
