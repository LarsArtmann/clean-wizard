// Package lockutil provides generic helpers for [sync.Mutex]- and
// [sync.RWMutex]-protected critical sections. It eliminates the
// m.mu.Lock()/defer m.mu.Unlock() boilerplate that otherwise accumulates
// around every short read or write on shared state.
//
// Usage:
//
//	type Counter struct {
//	    mu    sync.RWMutex
//	    value int
//	}
//
//	func (c *Counter) Get() int {
//	    return lockutil.RLocked(&c.mu, func() int { return c.value })
//	}
//
//	func (c *Counter) Inc() {
//	    lockutil.Locked(&c.mu, func() struct{} {
//	        c.value++
//	        return struct{}{}
//	    })
//	}
//
// The helpers return a generic T so callers can return values directly
// without intermediate variables. Use struct{} for side-effect-only
// critical sections. When a closure must return multiple values, wrap
// them in a small local struct (Go cannot infer T for tuple-shaped returns).
//
// Locked works with any [sync.Locker] (both [sync.Mutex] and
// [sync.RWMutex]). RLocked is specific to [*sync.RWMutex].
//
// This package depends only on Go stdlib.
package lockutil

import "sync"

// RLocked runs fn while holding mu.RLock and returns its result.
func RLocked[T any](mu *sync.RWMutex, fn func() T) T {
	mu.RLock()
	defer mu.RUnlock()

	return fn()
}

// Locked runs fn while holding the write lock and returns its result.
// Works with any [sync.Locker]; most callers pass &mu where mu is
// [sync.Mutex] or [sync.RWMutex].
func Locked[T any](mu sync.Locker, fn func() T) T {
	mu.Lock()
	defer mu.Unlock()

	return fn()
}
