package execution_test

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// fakeCleaner is a scriptable cleaner double for behavior specs. Outcomes are
// consumed in order per Clean call; the last outcome repeats for further calls.
type fakeCleaner struct {
	name       string
	available  bool
	outcomes   []error
	cleanDelay time.Duration
	calls      atomic.Int32

	concurrencyMu   sync.Mutex
	current         int32
	peakConcurrency int32
}

func newFakeCleaner(name string, outcomes ...error) *fakeCleaner {
	return &fakeCleaner{name: name, available: true, outcomes: outcomes}
}

func (f *fakeCleaner) Name() string                      { return f.name }
func (f *fakeCleaner) Type() domain.OperationType        { return domain.OperationTypeCargoPackages }
func (f *fakeCleaner) IsAvailable(_ context.Context) bool { return f.available }

func (f *fakeCleaner) Scan(_ context.Context) result.Result[[]domain.ScanItem] {
	return result.Ok([]domain.ScanItem{{Size: 100}, {Size: 200}}) //nolint:exhaustruct
}

func (f *fakeCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	call := f.calls.Add(1)

	f.concurrencyMu.Lock()
	f.current++
	if f.current > f.peakConcurrency {
		f.peakConcurrency = f.current
	}
	f.concurrencyMu.Unlock()

	defer func() {
		f.concurrencyMu.Lock()
		f.current--
		f.concurrencyMu.Unlock()
	}()

	if f.cleanDelay > 0 {
		select {
		case <-ctx.Done():
			return result.Err[domain.CleanResult](ctx.Err())
		case <-time.After(f.cleanDelay):
		}
	}

	outcomeIdx := int(call) - 1
	if outcomeIdx >= len(f.outcomes) {
		outcomeIdx = len(f.outcomes) - 1
	}

	if outcomeIdx >= 0 && f.outcomes[outcomeIdx] != nil {
		return result.Err[domain.CleanResult](f.outcomes[outcomeIdx])
	}

	return result.Ok(domain.CleanResult{FreedBytes: 100, ItemsRemoved: 1}) //nolint:exhaustruct
}

func (f *fakeCleaner) callCount() int32 { return f.calls.Load() }

func (f *fakeCleaner) peak() int32 {
	f.concurrencyMu.Lock()
	defer f.concurrencyMu.Unlock()

	return f.peakConcurrency
}

// registerFakes registers the given cleaners in order and returns the registry.
func registerFakes(cleaners ...*fakeCleaner) *cleaner.Registry {
	registry := cleaner.NewRegistry()
	for _, c := range cleaners {
		registry.Register(c.name, c)
	}

	return registry
}
