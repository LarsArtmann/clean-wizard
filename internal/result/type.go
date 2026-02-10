package result

import (
	"errors"
)

// Result is a type-safe way to return values or errors.
type Result[T any] struct {
	value T
	err   error
}

// Ok creates a successful result.
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value}
}

// Err creates an error result.
func Err[T any](err error) Result[T] {
	var zero T
	return Result[T]{value: zero, err: err}
}

// IsOk returns true if result is successful.
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// IsErr returns true if result has error.
func (r Result[T]) IsErr() bool {
	return r.err != nil
}

// Unwrap returns value and error.
func (r Result[T]) Unwrap() (T, error) {
	return r.value, r.err
}

// Value returns value (panics if error).
func (r Result[T]) Value() T {
	if r.err != nil {
		panic("attempted to get value from error result: " + r.err.Error())
	}
	return r.value
}

// SafeValue returns value and error (never panics).
func (r Result[T]) SafeValue() (T, error) {
	if r.err != nil {
		var zero T
		return zero, r.err
	}
	return r.value, nil
}

// Error returns error (panics on success).
func (r Result[T]) Error() error {
	if r.err == nil {
		panic("attempted to get error from success result")
	}
	return r.err
}

// SafeError returns error and ok boolean (never panics).
func (r Result[T]) SafeError() (error, bool) {
	return r.err, r.err != nil
}

// UnwrapOr returns value or default if error.
func (r Result[T]) UnwrapOr(default_ T) T {
	if r.err != nil {
		return default_
	}
	return r.value
}

// Map applies function to value if successful, passes through error.
func Map[T, U any](r Result[T], fn func(T) U) Result[U] {
	if r.err != nil {
		return Err[U](r.err)
	}
	return Ok(fn(r.value))
}

// AndThen chains operations that return Result, flattening the result.
// If the result is an error, it returns an error result without calling the function.
// This enables chaining operations that can fail (like FlatMap in functional programming).
func AndThen[T, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	if r.err != nil {
		return Err[U](r.err)
	}
	return fn(r.value)
}

// FlatMap is an alias for AndThen for semantic clarity when the function returns a Result.
func FlatMap[T, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	return AndThen(r, fn)
}

// OrElse returns the current result if successful, otherwise returns the fallback result.
// Useful for providing fallback values when operations fail.
func (r Result[T]) OrElse(fallback Result[T]) Result[T] {
	if r.err != nil {
		return fallback
	}
	return r
}

// Validate checks if the value satisfies the predicate.
// If the predicate returns false, it returns an error with the given message.
// If the result already has an error, it passes through the error.
func (r Result[T]) Validate(predicate func(T) bool, errorMsg string) Result[T] {
	if r.err != nil {
		return r
	}
	if !predicate(r.value) {
		return Err[T](errors.New(errorMsg))
	}
	return r
}

// ValidateWithError checks if the value satisfies the predicate.
// If the predicate returns false, it returns the provided error.
// If the result already has an error, it passes through the error.
func (r Result[T]) ValidateWithError(predicate func(T) bool, err error) Result[T] {
	if r.err != nil {
		return r
	}
	if !predicate(r.value) {
		return Err[T](err)
	}
	return r
}

// Tap applies a side-effect function to the value if successful.
// Returns the original result unchanged, useful for logging or other side effects.
func (r Result[T]) Tap(fn func(T)) Result[T] {
	if r.err == nil {
		fn(r.value)
	}
	return r
}

// MockSuccess creates a successful result with warning message.
func MockSuccess[T any](value T, message string) Result[T] {
	return Ok(value)
}
