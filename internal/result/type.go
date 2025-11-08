package result

// Result is a type-safe way to return values or errors
type Result[T any] struct {
	value T
	err   error
}

// Ok creates a successful result
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value}
}

// Err creates an error result
func Err[T any](err error) Result[T] {
	var zero T
	return Result[T]{value: zero, err: err}
}

// IsOk returns true if result is successful
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// IsErr returns true if result has error
func (r Result[T]) IsErr() bool {
	return r.err != nil
}

// Unwrap returns value and error
func (r Result[T]) Unwrap() (T, error) {
	return r.value, r.err
}

// Value returns value (panics if error)
func (r Result[T]) Value() T {
	if r.err != nil {
		panic("attempted to get value from error result: " + r.err.Error())
	}
	return r.value
}

// Error returns error (nil if success)
func (r Result[T]) Error() error {
	return r.err
}
