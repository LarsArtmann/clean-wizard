package result

import (
	"fmt"
	"time"
)

// Result represents the result of an operation that can fail
type Result[T any] struct {
	value T
	err   error
}

// Ok creates a successful result
func Ok[T any](value T) Result[T] {
	return Result[T]{
		value: value,
		err:   nil,
	}
}

// Err creates an error result
func Err[T any](err error) Result[T] {
	var zero T
	return Result[T]{
		value: zero,
		err:   err,
	}
}

// IsOk returns true if the result is successful
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// IsErr returns true if the result is an error
func (r Result[T]) IsErr() bool {
	return r.err != nil
}

// Value returns the value if successful, panics if error
func (r Result[T]) Value() T {
	if r.err != nil {
		panic("attempted to get value from error result")
	}
	return r.value
}

// ValueOr returns the value if successful, or default if error
func (r Result[T]) ValueOr(defaultValue T) T {
	if r.err != nil {
		return defaultValue
	}
	return r.value
}

// ValueOrFunc returns the value if successful, or calls function if error
func (r Result[T]) ValueOrFunc(fn func(error) T) T {
	if r.err != nil {
		return fn(r.err)
	}
	return r.value
}

// Error returns the error if error result, nil if successful
func (r Result[T]) Error() error {
	return r.err
}

// ErrorOr returns the error if error result, or default if successful
func (r Result[T]) ErrorOr(defaultError error) error {
	if r.err != nil {
		return r.err
	}
	return defaultError
}

// Unwrap returns the underlying error or panics if result is an error
func (r Result[T]) Unwrap() T {
	if r.err != nil {
		panic("attempted to unwrap error result")
	}
	return r.value
}

// UnwrapOr returns to value if successful, or default if error
func (r Result[T]) UnwrapOr(defaultValue T) T {
	if r.err != nil {
		return defaultValue
	}
	return r.value
}

// Match handles both success and error cases
func (r Result[T]) Match(
	onValue func(T) error,
	onError func(error) error,
) error {
	if r.err != nil {
		return onError(r.err)
	}
	return onValue(r.value)
}

// Map transforms the value if successful
func Map[T, U any](r Result[T], fn func(T) U) Result[U] {
	if r.err != nil {
		var zero U
		return Result[U]{
			value: zero,
			err:   r.err,
		}
	}
	return Ok(fn(r.value))
}

// MapError transforms the error if error result
func MapError[T any](r Result[T], fn func(error) error) Result[T] {
	if r.err != nil {
		return Err[T](fn(r.err))
	}
	return r
}

// FlatMap transforms the result to another result
func FlatMap[T, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	if r.err != nil {
		var zero U
		return Result[U]{
			value: zero,
			err:   r.err,
		}
	}
	return fn(r.value)
}

// AndThen chains operations, only executes next if successful
func AndThen[T, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	return FlatMap(r, fn)
}

// OrElse chains operations, only executes next if error
func OrElse[T any](r Result[T], fn func(error) Result[T]) Result[T] {
	if r.err != nil {
		return fn(r.err)
	}
	return r
}

// ResultToOptional converts Result to Optional
func ResultToOptional[T any](r Result[T]) Optional[T] {
	if r.err != nil {
		return Empty[T]()
	}
	return Some(r.value)
}

// Optional represents an optional value
type Optional[T any] struct {
	value T
	valid bool
}

// Some creates a non-empty optional
func Some[T any](value T) Optional[T] {
	return Optional[T]{
		value: value,
		valid: true,
	}
}

// Empty creates an empty optional
func Empty[T any]() Optional[T] {
	var zero T
	return Optional[T]{
		value: zero,
		valid: false,
	}
}

// IsPresent returns true if the optional has a value
func (o Optional[T]) IsPresent() bool {
	return o.valid
}

// IsEmpty returns true if the optional has no value
func (o Optional[T]) IsEmpty() bool {
	return !o.valid
}

// Value returns the value if present, panics if empty
func (o Optional[T]) Value() T {
	if !o.valid {
		panic("attempted to get value from empty optional")
	}
	return o.value
}

// ValueOr returns the value if present, or default if empty
func (o Optional[T]) ValueOr(defaultValue T) T {
	if o.valid {
		return o.value
	}
	return defaultValue
}

// ValueOrFunc returns the value if present, or calls function if empty
func (o Optional[T]) ValueOrFunc(fn func() T) T {
	if o.valid {
		return o.value
	}
	return fn()
}

// IfPresent executes function if value is present
func (o Optional[T]) IfPresent(fn func(T)) {
	if o.valid {
		fn(o.value)
	}
}

// IfPresentOrElse executes function if value is present, else another function
func (o Optional[T]) IfPresentOrElse(ifFn func(T), elseFn func()) {
	if o.valid {
		ifFn(o.value)
	} else {
		elseFn()
	}
}

// Map transforms the value if present
func MapOptional[T, U any](o Optional[T], fn func(T) Optional[U]) Optional[U] {
	if !o.valid {
		return Empty[U]()
	}
	return fn(o.value)
}

// Filter returns empty if predicate fails, else original
func (o Optional[T]) Filter(predicate func(T) bool) Optional[T] {
	if o.valid && predicate(o.value) {
		return o
	}
	return Empty[T]()
}

// AsyncResult represents an async operation result
type AsyncResult[T any] struct {
	result chan Result[T]
}

// NewAsyncResult creates new async result
func NewAsyncResult[T any]() *AsyncResult[T] {
	return &AsyncResult[T]{
		result: make(chan Result[T], 1),
	}
}

// Complete completes the async result with value
func (ar *AsyncResult[T]) Complete(value T) {
	ar.result <- Ok(value)
}

// CompleteError completes the async result with error
func (ar *AsyncResult[T]) CompleteError(err error) {
	ar.result <- Err[T](err)
}

// Wait waits for the result
func (ar *AsyncResult[T]) Wait() Result[T] {
	return <-ar.result
}

// WaitWithTimeout waits for result with timeout
func (ar *AsyncResult[T]) WaitWithTimeout(timeout time.Duration) Result[T] {
	select {
	case result := <-ar.result:
		return result
	case <-time.After(timeout):
		return Err[T](fmt.Errorf("operation timed out after %v", timeout))
	}
}

// Promise represents a value that will be available in the future
type Promise[T any] struct {
	AsyncResult[T]
}

// NewPromise creates new promise
func NewPromise[T any]() *Promise[T] {
	return &Promise[T]{
		AsyncResult: *NewAsyncResult[T](),
	}
}

// Resolve resolves the promise with value
func (p *Promise[T]) Resolve(value T) {
	p.Complete(value)
}

// Reject rejects the promise with error
func (p *Promise[T]) Reject(err error) {
	p.CompleteError(err)
}

// Then chains promise operations
func Then[T, U any](p *Promise[T], fn func(T) *Promise[U]) *Promise[U] {
	result := NewPromise[U]()

	go func() {
		r := p.Wait()
		if r.IsErr() {
			result.CompleteError(r.Error())
			return
		}

		nextPromise := fn(r.Value())
		nextResult := nextPromise.Wait()
		if nextResult.IsErr() {
			result.CompleteError(nextResult.Error())
		} else {
			result.Complete(nextResult.Value())
		}
	}()

	return result
}

// Catch handles promise errors
func Catch[T any](p *Promise[T], fn func(error) *Promise[T]) *Promise[T] {
	result := NewPromise[T]()

	go func() {
		r := p.Wait()
		if r.IsErr() {
			nextPromise := fn(r.Error())
			nextResult := nextPromise.Wait()
			if nextResult.IsErr() {
				result.CompleteError(nextResult.Error())
			} else {
				result.Complete(nextResult.Value())
			}
		} else {
			result.Complete(r.Value())
		}
	}()

	return result
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
	}
}

// Retry executes operation with retry logic
func Retry[T any](fn func() Result[T], config RetryConfig) Result[T] {
	var lastErr error

	for i := 0; i <= config.MaxRetries; i++ {
		result := fn()
		if result.IsOk() {
			return result
		}

		lastErr = result.Error()

		// Don't wait after last retry
		if i < config.MaxRetries {
			delay := calculateDelay(i, config)
			time.Sleep(delay)
		}
	}

	return Err[T](fmt.Errorf("operation failed after %d retries: %w", config.MaxRetries, lastErr))
}

// calculateDelay calculates delay for retry attempt
func calculateDelay(attempt int, config RetryConfig) time.Duration {
	delay := time.Duration(float64(config.InitialDelay) *
		float64(attempt) * config.BackoffFactor)

	if delay > config.MaxDelay {
		delay = config.MaxDelay
	}

	return delay
}

// ResultError represents a result error with context
type ResultError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details"`
	Cause   error                  `json:"cause"`
}

// Error implements error interface
func (re *ResultError) Error() string {
	return fmt.Sprintf("[%s] %s", re.Code, re.Message)
}

// Unwrap returns underlying cause
func (re *ResultError) Unwrap() error {
	return re.Cause
}

// NewResultError creates new result error
func NewResultError(code, message string, cause error) *ResultError {
	return &ResultError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
		Cause:   cause,
	}
}

// WithDetail adds detail to result error
func (re *ResultError) WithDetail(key string, value interface{}) *ResultError {
	re.Details[key] = value
	return re
}

// ResultBuilder helps build results with validation
type ResultBuilder[T any] struct {
	value T
	errs  []error
	warns []string
}

// NewResultBuilder creates new result builder
func NewResultBuilder[T any]() *ResultBuilder[T] {
	return &ResultBuilder[T]{
		errs:  make([]error, 0),
		warns: make([]string, 0),
	}
}

// SetValue sets the value
func (rb *ResultBuilder[T]) SetValue(value T) *ResultBuilder[T] {
	rb.value = value
	return rb
}

// AddError adds an error
func (rb *ResultBuilder[T]) AddError(err error) *ResultBuilder[T] {
	rb.errs = append(rb.errs, err)
	return rb
}

// AddWarning adds a warning
func (rb *ResultBuilder[T]) AddWarning(warning string) *ResultBuilder[T] {
	rb.warns = append(rb.warns, warning)
	return rb
}

// Validate validates condition
func (rb *ResultBuilder[T]) Validate(condition bool, err error) *ResultBuilder[T] {
	if !condition && err != nil {
		rb.errs = append(rb.errs, err)
	}
	return rb
}

// Build builds the result
func (rb *ResultBuilder[T]) Build() Result[T] {
	if len(rb.errs) > 0 {
		if len(rb.errs) == 1 {
			return Err[T](rb.errs[0])
		}
		return Err[T](fmt.Errorf("multiple errors: %v", rb.errs))
	}
	return Ok(rb.value)
}

// BuildWithWarnings builds result with warnings
func (rb *ResultBuilder[T]) BuildWithWarnings() (Result[T], []string) {
	return rb.Build(), rb.warns
}
