package finding

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/larsartmann/go-finding/lockutil"
)

// Sentinel errors for the detector registry. Exported so callers can use
// errors.Is to distinguish registration conflicts from unknown detectors.
var (
	ErrDetectorRegistered = errors.New("detector already registered")
	ErrUnknownDetector    = errors.New("unknown detector")

	// Deprecated aliases kept for internal backward compatibility.
	errDetectorRegistered = ErrDetectorRegistered
	errUnknownDetector    = ErrUnknownDetector
)

// DetectorRegistry manages named detector constructors.
// Create one with [NewDetectorRegistry] and register detectors
// with [DetectorRegistry.Register]. Use [DetectorRegistry.Build] to
// instantiate detectors by name.
//
// The registry is safe for concurrent use.
type DetectorRegistry struct {
	mu       sync.RWMutex
	builders map[string]func() Detector
}

// NewDetectorRegistry creates an empty registry.
func NewDetectorRegistry() *DetectorRegistry {
	return &DetectorRegistry{builders: make(map[string]func() Detector)}
}

// Register adds a detector constructor under the given name.
// Returns an error if a detector with the same name is already registered.
func (r *DetectorRegistry) Register(name string, builder func() Detector) error {
	return writeDetector(r, func() error {
		if _, exists := r.builders[name]; exists {
			return fmt.Errorf("%w: %s", errDetectorRegistered, name)
		}

		r.builders[name] = builder

		return nil
	})
}

// MustRegister panics if registration fails.
func (r *DetectorRegistry) MustRegister(name string, builder func() Detector) {
	err := r.Register(name, builder)
	if err != nil {
		panic(err)
	}
}

// Build instantiates a detector by name. Returns an error if not found.
//
//nolint:ireturn
func (r *DetectorRegistry) Build(name string) (Detector, error) {
	type lookup struct {
		builder func() Detector
		ok      bool
	}

	res := readDetector(r, func() lookup {
		b, ok := r.builders[name]

		return lookup{builder: b, ok: ok}
	})
	if !res.ok {
		return nil, fmt.Errorf("%w: %s", errUnknownDetector, name)
	}

	return res.builder(), nil
}

// BuildAll instantiates all registered detectors in sorted name order.
func (r *DetectorRegistry) BuildAll() ([]Detector, error) {
	names := readDetector(r, func() []string {
		return slices.Sorted(maps.Keys(r.builders))
	})

	detectors := make([]Detector, 0, len(names))

	for _, name := range names {
		d, err := r.Build(name)
		if err != nil {
			return nil, err
		}

		detectors = append(detectors, d)
	}

	return detectors, nil
}

// Names returns registered detector names in sorted order.
func (r *DetectorRegistry) Names() []string {
	return readDetector(r, func() []string {
		return slices.Sorted(maps.Keys(r.builders))
	})
}

// Has reports whether a detector with the given name is registered.
func (r *DetectorRegistry) Has(name string) bool {
	return readDetector(r, func() bool {
		_, ok := r.builders[name]

		return ok
	})
}

// readDetector runs fn while holding r.mu.RLock and returns its result.
func readDetector[T any](r *DetectorRegistry, fn func() T) T {
	return lockutil.RLocked(&r.mu, fn)
}

// writeDetector runs fn while holding r.mu.Lock and returns its result.
func writeDetector[T any](r *DetectorRegistry, fn func() T) T {
	return lockutil.Locked(&r.mu, fn)
}
