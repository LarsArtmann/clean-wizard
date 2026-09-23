package execution_test

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// concurrencyTracker measures how many fake cleaners overlap in time.
// It is shared by all cleaners in one registry so overlap is visible.
type concurrencyTracker struct {
	mu      sync.Mutex
	current int32
	peak    int32
}

func (t *concurrencyTracker) enter() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.current++
	if t.current > t.peak {
		t.peak = t.current
	}
}

func (t *concurrencyTracker) exit() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.current--
}

func (t *concurrencyTracker) Peak() int32 {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.peak
}

// fakeCleaner is a scriptable cleaner double for behavior specs. Outcomes are
// consumed in order per Clean call; the last outcome repeats for further calls.
type fakeCleaner struct {
	name       string
	available  bool
	outcomes   []error
	cleanDelay time.Duration
	tracker    *concurrencyTracker
	calls      atomic.Int32
}

func newFakeCleaner(name string, outcomes ...error) *fakeCleaner {
	return &fakeCleaner{name: name, available: true, outcomes: outcomes}
}

func (f *fakeCleaner) Name() string                       { return f.name }
func (f *fakeCleaner) Type() operations.OperationType     { return operations.OperationTypeCargoPackages }
func (f *fakeCleaner) IsAvailable(_ context.Context) bool { return f.available }

func (f *fakeCleaner) Scan(_ context.Context) result.Result[[]types.ScanItem] {
	return result.Ok([]types.ScanItem{{Size: 100}, {Size: 200}})
}

func (f *fakeCleaner) Clean(ctx context.Context) result.Result[types.CleanResult] {
	call := f.calls.Add(1)

	if f.tracker != nil {
		f.tracker.enter()
		defer f.tracker.exit()
	}

	if f.cleanDelay > 0 {
		select {
		case <-ctx.Done():
			return result.Err[types.CleanResult](ctx.Err())
		case <-time.After(f.cleanDelay):
		}
	}

	outcomeIdx := int(call) - 1
	if outcomeIdx >= len(f.outcomes) {
		outcomeIdx = len(f.outcomes) - 1
	}

	if outcomeIdx >= 0 && f.outcomes[outcomeIdx] != nil {
		return result.Err[types.CleanResult](f.outcomes[outcomeIdx])
	}

	return result.Ok(types.CleanResult{FreedBytes: 100, ItemsRemoved: 1})
}

func (f *fakeCleaner) callCount() int32 { return f.calls.Load() }

// registerFakes registers the given cleaners in order and returns the registry
// plus a concurrency tracker shared by all of them.
func registerFakes(cleaners ...*fakeCleaner) (*cleaner.Registry, *concurrencyTracker) {
	tracker := &concurrencyTracker{}
	registry := cleaner.NewRegistry()

	for _, c := range cleaners {
		c.tracker = tracker
		registry.Register(c.name, c)
	}

	return registry, tracker
}
