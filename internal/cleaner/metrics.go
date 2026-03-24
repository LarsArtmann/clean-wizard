package cleaner

import (
	"context"
	"sync"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// MetricsCollector provides observability hooks for cleaner operations.
// It can be used to track performance, success rates, and other metrics.
type MetricsCollector struct {
	mu       sync.RWMutex
	cleaners map[string]*CleanerMetrics
}

// CleanerMetrics tracks metrics for a specific cleaner.
type CleanerMetrics struct {
	Name            string
	InvocationCount uint64
	SuccessCount    uint64
	FailureCount    uint64
	TotalDuration   time.Duration
	TotalBytesFreed uint64
	LastRunAt       time.Time
	LastError       error
}

// MetricsSnapshot provides a point-in-time view of all metrics.
type MetricsSnapshot struct {
	Cleaners        map[string]CleanerMetrics
	TotalCleaners   int
	TotalRuns       uint64
	TotalSuccesses  uint64
	TotalFailures   uint64
	TotalBytesFreed uint64
}

// NewMetricsCollector creates a new metrics collector.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		cleaners: make(map[string]*CleanerMetrics),
	}
}

// RecordStart records the start of a cleaner operation.
func (mc *MetricsCollector) RecordStart(cleanerName string) time.Time {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	m, exists := mc.cleaners[cleanerName]
	if !exists {
		m = &CleanerMetrics{Name: cleanerName}
		mc.cleaners[cleanerName] = m
	}

	m.InvocationCount++

	return time.Now()
}

// RecordSuccess records a successful cleaner operation.
func (mc *MetricsCollector) RecordSuccess(
	cleanerName string,
	startTime time.Time,
	res domain.CleanResult,
) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	m, exists := mc.cleaners[cleanerName]
	if !exists {
		m = &CleanerMetrics{Name: cleanerName}
		mc.cleaners[cleanerName] = m
	}

	duration := time.Since(startTime)

	m.SuccessCount++
	m.TotalDuration += duration
	m.TotalBytesFreed += res.FreedBytes
	m.LastRunAt = time.Now()
}

// RecordFailure records a failed cleaner operation.
func (mc *MetricsCollector) RecordFailure(
	cleanerName string,
	startTime time.Time,
	err error,
) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	m, exists := mc.cleaners[cleanerName]
	if !exists {
		m = &CleanerMetrics{Name: cleanerName}
		mc.cleaners[cleanerName] = m
	}

	m.FailureCount++
	m.LastError = err
	m.LastRunAt = time.Now()
}

// GetMetrics returns metrics for a specific cleaner.
func (mc *MetricsCollector) GetMetrics(cleanerName string) (CleanerMetrics, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	m, ok := mc.cleaners[cleanerName]
	if !ok {
		return CleanerMetrics{}, false
	}

	return *m, true
}

// Snapshot returns a snapshot of all metrics.
func (mc *MetricsCollector) Snapshot() MetricsSnapshot {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	snapshot := MetricsSnapshot{
		Cleaners:      make(map[string]CleanerMetrics, len(mc.cleaners)),
		TotalCleaners: len(mc.cleaners),
	}

	for name, m := range mc.cleaners {
		snapshot.Cleaners[name] = *m
		snapshot.TotalRuns += m.InvocationCount
		snapshot.TotalSuccesses += m.SuccessCount
		snapshot.TotalFailures += m.FailureCount
		snapshot.TotalBytesFreed += m.TotalBytesFreed
	}

	return snapshot
}

// Reset clears all metrics.
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.cleaners = make(map[string]*CleanerMetrics)
}

// TrackedCleaner wraps a cleaner with metrics collection.
type TrackedCleaner struct {
	Cleaner

	collector *MetricsCollector
}

// NewTrackedCleaner creates a new tracked cleaner wrapper.
func NewTrackedCleaner(c Cleaner, collector *MetricsCollector) *TrackedCleaner {
	return &TrackedCleaner{
		Cleaner:   c,
		collector: collector,
	}
}

// Clean executes the cleaner and records metrics.
func (tc *TrackedCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	start := tc.collector.RecordStart(tc.Name())
	res := tc.Cleaner.Clean(ctx)

	if res.IsOk() {
		result, _ := res.Unwrap()
		tc.collector.RecordSuccess(tc.Name(), start, result)
	} else {
		_, err := res.Unwrap()
		tc.collector.RecordFailure(tc.Name(), start, err)
	}

	return res
}

// MetricsEnabledRegistry extends Registry with metrics collection.
type MetricsEnabledRegistry struct {
	*Registry

	collector *MetricsCollector
}

// NewMetricsEnabledRegistry creates a new metrics-enabled registry.
func NewMetricsEnabledRegistry(registry *Registry) *MetricsEnabledRegistry {
	return &MetricsEnabledRegistry{
		Registry:  registry,
		collector: NewMetricsCollector(),
	}
}

// GetCollector returns the metrics collector.
func (r *MetricsEnabledRegistry) GetCollector() *MetricsCollector {
	return r.collector
}

// CleanAllWithMetrics runs all cleaners with metrics collection.
func (r *MetricsEnabledRegistry) CleanAllWithMetrics(
	ctx context.Context,
) map[string]result.Result[domain.CleanResult] {
	available := r.Available(ctx)
	results := make(map[string]result.Result[domain.CleanResult], len(available))

	for _, c := range available {
		tracked := NewTrackedCleaner(c, r.collector)
		results[c.Name()] = tracked.Clean(ctx)
	}

	return results
}

// CleanAllParallelWithMetrics runs all cleaners in parallel with metrics.
func (r *MetricsEnabledRegistry) CleanAllParallelWithMetrics(
	ctx context.Context,
	maxConcurrency int,
) map[string]CleanResultWithMetrics {
	available := r.Available(ctx)
	executor := NewParallelExecutor(maxConcurrency)

	// Wrap cleaners with tracking
	trackedCleaners := make([]Cleaner, len(available))
	for i, c := range available {
		trackedCleaners[i] = NewTrackedCleaner(c, r.collector)
	}

	results := executor.Execute(ctx, trackedCleaners)

	// Convert to map
	resultMap := make(map[string]CleanResultWithMetrics, len(results))
	for _, res := range results {
		resultMap[res.Cleaner] = res
	}

	return resultMap
}
