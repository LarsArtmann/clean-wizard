package cleaner

import (
	"context"
	"sync"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// CleanResultWithMetrics wraps a clean result with execution metrics.
type CleanResultWithMetrics struct {
	Result   result.Result[domain.CleanResult]
	Duration time.Duration
	Cleaner  string
}

// ParallelExecutor executes cleaners in parallel with configurable concurrency.
type ParallelExecutor struct {
	maxConcurrency int
	mu             sync.RWMutex
}

// NewParallelExecutor creates a new parallel executor with the specified max concurrency.
// If maxConcurrency <= 0, it defaults to the number of available cleaners.
func NewParallelExecutor(maxConcurrency int) *ParallelExecutor {
	return &ParallelExecutor{
		maxConcurrency: maxConcurrency,
	}
}

// SetMaxConcurrency updates the maximum concurrency level.
func (pe *ParallelExecutor) SetMaxConcurrency(maxConcurrency int) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	pe.maxConcurrency = maxConcurrency
}

// GetMaxConcurrency returns the current maximum concurrency level.
func (pe *ParallelExecutor) GetMaxConcurrency() int {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	return pe.maxConcurrency
}

// Execute runs the given cleaners in parallel and returns results with metrics.
// The context can be used to cancel all ongoing operations.
func (pe *ParallelExecutor) Execute(
	ctx context.Context,
	cleaners []Cleaner,
) []CleanResultWithMetrics {
	if len(cleaners) == 0 {
		return []CleanResultWithMetrics{}
	}

	pe.mu.RLock()
	maxConcurrency := pe.maxConcurrency
	pe.mu.RUnlock()

	if maxConcurrency <= 0 {
		maxConcurrency = len(cleaners)
	}

	// Limit concurrency to number of cleaners
	if maxConcurrency > len(cleaners) {
		maxConcurrency = len(cleaners)
	}

	results := make([]CleanResultWithMetrics, len(cleaners))

	// Use semaphore pattern for concurrency control
	semaphore := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup

	for i, c := range cleaners {
		wg.Add(1)

		go func(index int, cleaner Cleaner) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case semaphore <- struct{}{}:
				// Acquired
			case <-ctx.Done():
				// Context cancelled while waiting
				results[index] = CleanResultWithMetrics{
					Result:   result.Err[domain.CleanResult](ctx.Err()),
					Duration: 0,
					Cleaner:  cleaner.Name(),
				}

				return
			}

			defer func() { <-semaphore }() // Release semaphore

			start := time.Now()
			res := cleaner.Clean(ctx)
			duration := time.Since(start)

			results[index] = CleanResultWithMetrics{
				Result:   res,
				Duration: duration,
				Cleaner:  cleaner.Name(),
			}
		}(i, c)
	}

	wg.Wait()

	return results
}

// CleanAllParallel runs all available cleaners in parallel with configurable concurrency.
// This is a convenience method on Registry.
func (r *Registry) CleanAllParallel(
	ctx context.Context,
	maxConcurrency int,
) map[string]CleanResultWithMetrics {
	available := r.Available(ctx)
	executor := NewParallelExecutor(maxConcurrency)
	results := executor.Execute(ctx, available)

	// Convert to map
	resultMap := make(map[string]CleanResultWithMetrics, len(results))
	for _, res := range results {
		resultMap[res.Cleaner] = res
	}

	return resultMap
}

// ExecutionMetrics provides aggregated metrics for parallel execution.
type ExecutionMetrics struct {
	TotalCleaners   int
	Successful      int
	Failed          int
	TotalDuration   time.Duration
	TotalBytesFreed int64
}

// CalculateMetrics aggregates metrics from parallel execution results.
func CalculateMetrics(results []CleanResultWithMetrics) ExecutionMetrics {
	metrics := ExecutionMetrics{
		TotalCleaners: len(results),
	}

	for _, res := range results {
		metrics.TotalDuration += res.Duration

		if res.Result.IsOk() {
			metrics.Successful++
			result, _ := res.Result.Unwrap()
			metrics.TotalBytesFreed += int64(result.FreedBytes)
		} else {
			metrics.Failed++
		}
	}

	return metrics
}
