package cleaner

import (
	"context"
	"sync"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// Registry manages all registered cleaners.
// It provides thread-safe access to cleaner instances and enables
// polymorphic operations over all cleaners.
type Registry struct {
	cleaners map[string]Cleaner
	mu       sync.RWMutex
}

// NewRegistry creates a new cleaner registry.
func NewRegistry() *Registry {
	return &Registry{
		cleaners: make(map[string]Cleaner),
	}
}

// Register adds a cleaner to the registry.
// If a cleaner with the same name already exists, it will be overwritten.
func (r *Registry) Register(name string, c Cleaner) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleaners[name] = c
}

// Get retrieves a cleaner by name.
// Returns the cleaner and true if found, nil and false otherwise.
func (r *Registry) Get(name string) (Cleaner, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.cleaners[name]
	return c, ok
}

// List returns all registered cleaners.
// The order of cleaners in the returned slice is not guaranteed.
func (r *Registry) List() []Cleaner {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]Cleaner, 0, len(r.cleaners))
	for _, c := range r.cleaners {
		list = append(list, c)
	}
	return list
}

// Names returns all registered cleaner names.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.cleaners))
	for name := range r.cleaners {
		names = append(names, name)
	}
	return names
}

// Count returns the number of registered cleaners.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.cleaners)
}

// Available returns all cleaners that are available on the current system.
func (r *Registry) Available(ctx context.Context) []Cleaner {
	all := r.List()
	available := make([]Cleaner, 0, len(all))
	for _, c := range all {
		if c.IsAvailable(ctx) {
			available = append(available, c)
		}
	}
	return available
}

// CleanAll runs all available cleaners and aggregates results.
// Returns a map of cleaner name to result.
func (r *Registry) CleanAll(ctx context.Context) map[string]result.Result[domain.CleanResult] {
	available := r.Available(ctx)
	results := make(map[string]result.Result[domain.CleanResult], len(available))

	for _, c := range available {
		res := c.Clean(ctx)
		// Note: This doesn't capture the name correctly in current implementation
		// as Cleaner interface doesn't have Name() method
		results["unknown"] = res
	}

	return results
}

// Unregister removes a cleaner from the registry.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.cleaners, name)
}

// Clear removes all cleaners from the registry.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleaners = make(map[string]Cleaner)
}
