package errorfamily

import "errors"

// Classifier is a predicate-based error classifier for dynamic third-party errors
// that cannot be registered as sentinels — each error is a fresh instance not
// matchable by errors.Is identity (e.g. *sqlite.Error, *pgconn.PgError).
//
// Register one via [RegisterClassifier] or [Registry.RegisterClassifier]. During
// [Classify], after the Classified interface, Retryable inference, and registered
// sentinels all miss, each registered Classifier runs in registration order; the
// first to return ok=true wins. Classifiers only run on otherwise-unclassified
// errors, so the hot path (errors that already declare a family) is untouched.
//
// Example — classifying modernc.org/sqlite dynamic errors:
//
//	errorfamily.RegisterClassifier(func(err error) (errorfamily.Family, bool) {
//	    var sqliteErr *sqlite.Error
//	    if errors.As(err, &sqliteErr) {
//	        switch sqliteErr.Code() {
//	        case 5, 6: return errorfamily.Transient, true // BUSY, LOCKED
//	        case 19:   return errorfamily.Conflict, true  // CONSTRAINT
//	        }
//	    }
//	    return errorfamily.Transient, false
//	})
type Classifier func(error) (Family, bool)

// Code extracts the machine-readable error code from any error in the chain.
// It walks the unwrap chain via errors.AsType looking for the [Coded] interface.
// Returns "" if the error (or any wrapped cause) does not implement Coded.
//
// This is the one-liner replacement for:
//
//	var coded errorfamily.Coded
//	if errors.As(err, &coded) {
//	    code = coded.ErrorCode()
//	}
func Code(err error) string {
	if coded, ok := errors.AsType[Coded](err); ok {
		return coded.ErrorCode()
	}

	return ""
}

// Classify returns the Family of any error by checking multiple sources:
//
//  1. Multi-error support (errors.Join) — first non-Transient wins
//  2. Classified interface — the error itself declares its family
//  3. Retryable interface — infer from retryability if no family
//  4. Registered sentinels — known third-party errors mapped via RegisterClassification
//  5. Registered classifiers — predicate-based matching for dynamic errors (RegisterClassifier)
//  6. Default — Transient (fail-open for retry)
//
// Returns Rejection for nil errors.
//
// This function delegates to [DefaultRegistry]. For scoped or test-isolated
// classification, construct a [Registry] via [NewRegistry] and use its
// Classify method.
func Classify(err error) Family {
	return DefaultRegistry.Classify(err)
}

// IsRetryable reports whether the error is worth retrying.
// Uses Classify() and checks if the result is Transient.
func IsRetryable(err error) bool {
	return Classify(err).IsRetryable()
}

// ExitCode returns the appropriate process exit code for an error.
// Nil errors return 0.
//
// If the error (or any error in its chain) implements [ExitCoder] with a non-zero
// code, that code is returned — allowing individual errors to override their
// family's canonical exit code. Otherwise, the error is classified and mapped to
// BSD sysexits.h codes via [Family.ExitCode].
func ExitCode(err error) int {
	if err == nil {
		return 0
	}

	if ec, ok := errors.AsType[ExitCoder](err); ok {
		if code := ec.ExitCode(); code != 0 {
			return code
		}
	}

	return Classify(err).ExitCode()
}

// Compose joins multiple errors into one, preserving all in the Unwrap chain.
// It is a thin wrapper around [errors.Join] kept for backward compatibility
// with consumers that imported Compose before v0.5.0 reorganized the package.
func Compose(
	errs ...error,
) error {
	return errors.Join(errs...)
}

// RegisterClassification maps a third-party sentinel error to a Family.
// Thread-safe. Call from init() in external packages:
//
//	func init() {
//	    errorfamily.RegisterClassification(sql.ErrConnDone, errorfamily.Transient)
//	}
//
// This is for errors you don't own (stdlib, libraries).
// For your own errors, implement the Classified interface instead.
//
// Delegates to [DefaultRegistry]. For scoped registration, use
// [Registry.RegisterClassification] on a custom Registry.
func RegisterClassification(sentinel error, family Family) {
	DefaultRegistry.RegisterClassification(sentinel, family)
}

// UnregisterClassification removes a previously registered sentinel mapping.
// Thread-safe. No-op if the sentinel has no registered classification.
func UnregisterClassification(sentinel error) {
	DefaultRegistry.UnregisterClassification(sentinel)
}

// RegisterClassifications registers multiple sentinel-to-Family mappings at once.
// Thread-safe. Call from init() in external packages.
func RegisterClassifications(classifications map[error]Family) {
	DefaultRegistry.RegisterClassifications(classifications)
}

// RegisterClassifier adds a predicate-based [Classifier] for dynamic third-party
// errors that cannot be registered as sentinels (e.g. *sqlite.Error).
// Thread-safe. Classifiers run after sentinel matching fails, in registration
// order; the first returning ok=true wins.
//
// Because Go func values are not comparable, individual classifiers cannot be
// unregistered. For test isolation or scoped handling, construct a [NewRegistry]
// and call [Registry.RegisterClassifier] instead of polluting [DefaultRegistry].
//
// Delegates to [DefaultRegistry].
func RegisterClassifier(c Classifier) {
	DefaultRegistry.RegisterClassifier(c)
}

// RegisterClassifiers adds multiple predicate-based [Classifier] funcs at once.
// Thread-safe. Classifiers run in registration order. See [RegisterClassifier].
//
// Delegates to [DefaultRegistry].
func RegisterClassifiers(cs ...Classifier) {
	DefaultRegistry.RegisterClassifiers(cs...)
}

// RegisterClassificationType registers a type-based classifier that maps any
// error matching type T to the given Family. It is syntactic sugar over
// [RegisterClassifier] for the common case where the classification rule is
// simply "type T → Family F", without field-level inspection.
//
//	errorfamily.RegisterClassificationType[*sqlite.Error](errorfamily.Transient)
//
// For errors that need field-level logic (e.g. different sqlite error codes
// mapping to different families), use [RegisterClassifier] with a closure.
//
// For a custom [Registry], use [RegisterClassificationTypeFor].
//
// Delegates to [DefaultRegistry].
func RegisterClassificationType[T error](
	family Family,
) {
	RegisterClassificationTypeFor[T](DefaultRegistry, family)
}

// RegisterClassificationTypeFor registers a type-based classifier on a specific
// [Registry]. It is the generic counterpart to [Registry.RegisterClassifier]:
//
//	errorfamily.RegisterClassificationTypeFor[*sqlite.Error](myRegistry, errorfamily.Transient)
//
// Go does not permit type parameters on methods, so this is a top-level
// function rather than a method on [Registry].
func RegisterClassificationTypeFor[T error](
	r *Registry,
	family Family,
) {
	r.RegisterClassifier(func(err error) (Family, bool) {
		if _, ok := errors.AsType[T](
			err,
		); ok { //nolint:legacyerrors // presence check only: the typed value is intentionally unused, ok is checked
			return family, true
		}

		return Rejection, false
	})
}
