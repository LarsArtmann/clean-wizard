package errorfamily

import (
	"errors"
	"maps"
	"strings"
	"sync"
	"sync/atomic"
)

// sentinelMap is the immutable snapshot stored behind an atomic.Pointer.
// Reads load the pointer once (lock-free, allocation-free); writes copy-on-write.
type sentinelMap map[error]Family

// Registry holds classification sentinels and message templates.
// It is safe for concurrent use. The zero value is not usable — use [NewRegistry].
//
// A Registry solves two problems that package-level globals cannot:
//   - Test isolation: each test can construct its own Registry without polluting
//     global state or needing t.Cleanup(Unregister…) boilerplate.
//   - Scoped overrides: a service can register sentinels or templates for its
//     own error handling without affecting other parts of the same binary.
//
// The package-level [DefaultRegistry] is used by the convenience functions
// ([Classify], [RegisterClassification], [RegisterTemplate], etc.) and by the
// HandleError family when no custom Registry is provided in [HandleConfig].
//
// Sentinels are stored behind an atomic.Pointer to an immutable snapshot: the
// hot read path (every Classify) is lock-free and allocation-free, while rare
// writes serialize under the write lock and publish a new snapshot via copy-on-write.
type Registry struct {
	mu          sync.RWMutex // serializes writers (templates direct + sentinels/classifiers copy-on-write)
	sentinels   atomic.Pointer[sentinelMap]
	classifiers atomic.Pointer[[]Classifier]
	templates   map[string]MessageTemplate
}

// NewRegistry creates an empty Registry ready for use.
func NewRegistry() *Registry {
	reg := &Registry{
		templates: make(map[string]MessageTemplate),
	}
	empty := make(sentinelMap)
	reg.sentinels.Store(&empty)

	emptyC := make([]Classifier, 0)
	reg.classifiers.Store(&emptyC)

	return reg
}

// DefaultRegistry is the package-level Registry used by the convenience
// functions (Classify, RegisterClassification, RegisterTemplate, etc.) and by
// HandleError when HandleConfig.Registry is nil.
//
//nolint:gochecknoglobals // Package-level default registry for backward-compatible API.
var DefaultRegistry = NewRegistry()

// Classify returns the Family of any error using this registry's sentinel mappings.
//
// Checks in order, first match wins:
//  1. Multi-error (errors.Join) — first non-Transient sub-error wins
//  2. Classified interface — the error itself declares its family
//  3. Retryable interface — infer from retryability
//  4. Registered sentinels in this registry
//  5. Registered classifiers (predicate-based, for dynamic third-party errors)
//  6. Default — Transient (fail-open for retry)
//
// Returns Rejection for nil errors.
func (r *Registry) Classify(err error) Family {
	if err == nil {
		return Rejection
	}

	// 1. Multi-error support (errors.Join).
	// Pick the worst (highest-severity) sub-error, independent of join order.
	// Preserves fail-closed retry semantics: any non-Transient sub-error
	// (severity > Transient's) makes the joined result non-Transient.
	if u, ok := err.(interface{ Unwrap() []error }); ok {
		worst := Transient
		for _, sub := range u.Unwrap() {
			if f := r.Classify(sub); f.Severity() > worst.Severity() {
				worst = f
			}
		}

		return worst
	}

	// 2. Check for explicit classification.
	if c, ok := errors.AsType[Classified](err); ok {
		return c.ErrorFamily()
	}

	// 3. Check for retryability (infer family).
	if rl, ok := errors.AsType[Retryable](err); ok {
		if rl.IsRetryable() {
			return Transient
		}

		return Rejection
	}

	// 4. Check registered third-party sentinels.
	if family, ok := r.lookupSentinel(err); ok {
		return family
	}

	// 5. Run registered classifiers (for dynamic third-party errors).
	if family, ok := r.runClassifiers(err); ok {
		return family
	}

	// 6. Default: Transient (fail-open so unknown errors get retried).
	return Transient
}

// RegisterClassification maps a third-party sentinel error to a Family.
// Thread-safe. Call from init() in external packages:
//
//	func init() {
//	    errorfamily.DefaultRegistry.RegisterClassification(sql.ErrConnDone, errorfamily.Transient)
//	}
//
// This is for errors you don't own (stdlib, libraries).
// For your own errors, implement the Classified interface instead.
func (r *Registry) RegisterClassification(sentinel error, family Family) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.swapSentinels(func(m sentinelMap) { m[sentinel] = family })
}

// UnregisterClassification removes a previously registered sentinel mapping.
// Thread-safe. No-op if the sentinel has no registered classification.
func (r *Registry) UnregisterClassification(sentinel error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.swapSentinels(func(m sentinelMap) { delete(m, sentinel) })
}

// RegisterClassifications registers multiple sentinel-to-Family mappings at once.
// Thread-safe.
func (r *Registry) RegisterClassifications(classifications map[error]Family) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.swapSentinels(func(m sentinelMap) { maps.Copy(m, classifications) })
}

// RegisterClassifier adds a predicate-based [Classifier] for dynamic third-party
// errors that cannot be registered as sentinels (e.g. *sqlite.Error).
// Thread-safe. Classifiers run after sentinel matching fails, in registration
// order; the first returning ok=true wins.
//
// Because Go func values are not comparable, individual classifiers cannot be
// unregistered. For test isolation or scoped handling, construct a [NewRegistry]
// rather than polluting [DefaultRegistry].
func (r *Registry) RegisterClassifier(c Classifier) {
	r.RegisterClassifiers(c)
}

// RegisterClassifiers adds multiple predicate-based [Classifier] funcs at once.
// Thread-safe. Classifiers run in registration order.
func (r *Registry) RegisterClassifiers(classifiers ...Classifier) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.swapClassifiers(func(existing []Classifier) []Classifier {
		return append(existing, classifiers...)
	})
}

// swapClassifiers performs copy-on-write on the immutable classifier slice:
// clones the current snapshot, applies fn, and atomically publishes the result.
// Callers must hold r.mu (write lock) so concurrent writers never lose updates;
// readers remain fully lock-free.
func (r *Registry) swapClassifiers(fn func(cs []Classifier) []Classifier) {
	old := r.classifiers.Load()

	var newList []Classifier
	if old != nil {
		newList = make([]Classifier, len(*old), len(*old)+1)
		copy(newList, *old)
	}

	newList = fn(newList)
	r.classifiers.Store(&newList)
}

// runClassifiers loads the immutable classifier snapshot once (lock-free) and
// runs each classifier in registration order; the first ok=true wins.
func (r *Registry) runClassifiers(err error) (Family, bool) {
	cs := r.classifiers.Load()
	if cs == nil {
		return Rejection, false
	}

	for _, c := range *cs {
		if family, ok := c(err); ok {
			return family, true
		}
	}

	return Rejection, false
}

// swapSentinels performs copy-on-write: clones the current immutable sentinel
// snapshot, applies fn to the clone, and atomically publishes it. Callers must
// hold r.mu (write lock) so concurrent writers never lose updates; readers
// remain fully lock-free.
func (r *Registry) swapSentinels(fn func(m sentinelMap)) {
	old := r.sentinels.Load()

	newMap := make(sentinelMap, 1)
	if old != nil {
		newMap = make(sentinelMap, len(*old)+1)
		maps.Copy(newMap, *old)
	}

	fn(newMap)
	r.sentinels.Store(&newMap)
}

// RegisterTemplate adds a MessageTemplate for a specific error code.
// Thread-safe. Overrides any existing template for the same code.
func (r *Registry) RegisterTemplate(code string, tmpl MessageTemplate) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.templates[strings.ToLower(code)] = tmpl
}

// UnregisterTemplate removes a previously registered template.
// Thread-safe. No-op if the code has no registered template.
func (r *Registry) UnregisterTemplate(code string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.templates, strings.ToLower(code))
}

// RegisterTemplates registers multiple code-to-template mappings at once.
// Thread-safe. Keys are lower-cased for case-insensitive lookup.
// Parity with [Registry.RegisterClassifications] for the sentinel side.
func (r *Registry) RegisterTemplates(templates map[string]MessageTemplate) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for code, tmpl := range templates {
		r.templates[strings.ToLower(code)] = tmpl
	}
}

// Clone returns a new Registry with a deep copy of this Registry's sentinels
// and templates. Mutations to the clone do not affect the original and vice
// versa. Enables inherit-and-extend patterns: start from [DefaultRegistry],
// clone it, and register scope-specific overrides without touching the global.
func (r *Registry) Clone() *Registry {
	clone := NewRegistry()

	// Copy sentinels from the current immutable snapshot (lock-free read).
	if cur := r.sentinels.Load(); cur != nil {
		copied := make(sentinelMap, len(*cur))
		maps.Copy(copied, *cur)
		clone.sentinels.Store(&copied)
	}

	// Copy classifiers from the current immutable snapshot (lock-free read).
	if cur := r.classifiers.Load(); cur != nil {
		copied := make([]Classifier, 0, len(*cur))
		copied = append(copied, *cur...)
		clone.classifiers.Store(&copied)
	}

	// Copy templates under the read lock.
	r.mu.RLock()
	maps.Copy(clone.templates, r.templates)
	r.mu.RUnlock()

	return clone
}

// lookupSentinel loads the immutable sentinel snapshot once (lock-free,
// allocation-free) and walks the error chain against it.
func (r *Registry) lookupSentinel(err error) (Family, bool) {
	m := r.sentinels.Load()
	if m == nil {
		return Rejection, false
	}

	for sentinel, family := range *m {
		if errors.Is(err, sentinel) {
			return family, true
		}
	}

	return Rejection, false
}

// lookupTemplate looks up a user-registered template by code (case-insensitive).
func (r *Registry) lookupTemplate(code string) (MessageTemplate, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tmpl, ok := r.templates[strings.ToLower(code)]

	return tmpl, ok
}

// TemplateForCode resolves a [MessageTemplate] for an error code using this
// registry, checking registered templates first, then built-in defaults.
// Returns (zero, false) when no template exists for the code.
//
// This is the public counterpart to the internal resolution used by
// [HandleError]: it lets HTTP/REST consumers look up a registered template
// without reimplementing the lookup or wiring the full CLI pipeline.
func (r *Registry) TemplateForCode(code string) (MessageTemplate, bool) {
	if tmpl, ok := r.lookupTemplate(code); ok {
		return tmpl, true
	}

	if tmpl, ok := lookupDefault(code); ok {
		return tmpl, true
	}

	return MessageTemplate{}, false
}
