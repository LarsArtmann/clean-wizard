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

// Error returns error (panics on success)
func (r Result[T]) Error() error {
	if r.err == nil {
		panic("attempted to get error from success result")
	}
	return r.err
}

// UnwrapOr returns value or default if error
func (r Result[T]) UnwrapOr(default_ T) T {
	if r.err != nil {
		return default_
	}
	return r.value
}

// Map applies function to value if successful, passes through error
func Map[T, U any](r Result[T], fn func(T) U) Result[U] {
	if r.err != nil {
		return Err[U](r.err)
	}
	return Ok(fn(r.value))
}
