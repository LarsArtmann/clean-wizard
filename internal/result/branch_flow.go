// Package result provides functional programming patterns for error handling and flow control.
//
// This package offers comprehensive branching-flow context patterns to enhance the existing
// Result[T] type with pattern matching, conditional execution, pipeline composition, and
// parallel execution capabilities.
//
// Key types:
//   - Result[T]: Type-safe wrapper for values that may contain errors
//   - BranchFlow[T]: Complex conditional branching with fallback support
//   - FlowBuilder[T]: Sequential pipeline composition
//   - ParallelFlow[T]: Concurrent execution with result tracking
//
// Example usage:
//
//	flow := NewBranchFlow[int]().
//	    Branch(func() bool { return isAdmin }, func() Result[int] { return Ok(42) }).
//	    Fallback(func() Result[int] { return Ok(1) })
//
//	result := flow.Execute()
package result

import (
	"context"
)

// Branch represents a conditional branch in a BranchFlow.
type Branch[T any] struct {
	Condition func() bool      // Condition that determines if this branch is taken
	Execute   func() Result[T] // Function to execute if condition is true
}

// BranchFlow enables complex conditional branching flows with fallback support.
// It allows chaining multiple conditional branches that are evaluated in order,
// with an optional fallback for when no branch matches.
type BranchFlow[T any] struct {
	branches  []Branch[T]
	fallback  func() Result[T]
	finalizer func(Result[T]) Result[T]
}

// NewBranchFlow creates a new BranchFlow.
func NewBranchFlow[T any]() *BranchFlow[T] {
	return &BranchFlow[T]{
		branches:  []Branch[T]{},
		fallback:  nil,
		finalizer: nil,
	}
}

// Branch adds a conditional branch to the flow.
// The condition is evaluated first; if true, Execute is called and its result is returned.
func (bf *BranchFlow[T]) Branch(condition func() bool, execute func() Result[T]) *BranchFlow[T] {
	bf.branches = append(bf.branches, Branch[T]{
		Condition: condition,
		Execute:   execute,
	})

	return bf
}

// BranchWithValue adds a conditional branch that evaluates a predicate on a value.
// This is useful when you have a value and want to branch based on its properties.
func (bf *BranchFlow[T]) BranchWithValue(
	value T,
	condition func(T) bool,
	execute func(T) Result[T],
) *BranchFlow[T] {
	return bf.Branch(
		func() bool { return condition(value) },
		func() Result[T] { return execute(value) },
	)
}

// BranchWithContext adds a conditional branch with access to context.
// Useful for async or cancellable operations.
func (bf *BranchFlow[T]) BranchWithContext(
	ctx context.Context,
	condition func(context.Context) bool,
	execute func(context.Context) Result[T],
) *BranchFlow[T] {
	return bf.Branch(
		func() bool { return condition(ctx) },
		func() Result[T] { return execute(ctx) },
	)
}

// Fallback sets the fallback function to execute when no branch condition is true.
func (bf *BranchFlow[T]) Fallback(execute func() Result[T]) *BranchFlow[T] {
	bf.fallback = execute

	return bf
}

// FallbackValue provides a default value when no branch matches.
func (bf *BranchFlow[T]) FallbackValue(value T) *BranchFlow[T] {
	bf.fallback = func() Result[T] { return Ok(value) }

	return bf
}

// Finalize adds a finalizer function that transforms the result before returning.
// This is useful for adding logging, metrics, or additional validation.
func (bf *BranchFlow[T]) Finalize(fn func(Result[T]) Result[T]) *BranchFlow[T] {
	bf.finalizer = fn

	return bf
}

// Execute evaluates branches in order and returns the first matching result.
// If no branch matches and a fallback is set, executes the fallback.
// If no fallback is set, returns an error indicating no branch matched.
func (bf *BranchFlow[T]) Execute() Result[T] {
	// Evaluate branches in order
	for _, branch := range bf.branches {
		if branch.Condition() {
			result := branch.Execute()

			// Apply finalizer if set
			if bf.finalizer != nil {
				return bf.finalizer(result)
			}

			return result
		}
	}

	// No branch matched - execute fallback if set
	if bf.fallback != nil {
		result := bf.fallback()

		// Apply finalizer if set
		if bf.finalizer != nil {
			return bf.finalizer(result)
		}

		return result
	}

	// No fallback - return error
	return Err[T](ErrNoBranchMatched)
}

// ErrNoBranchMatched is returned when no branch condition was satisfied and no fallback was provided.
var ErrNoBranchMatched = &NoBranchMatchedError{}

// NoBranchMatchedError indicates that no branch condition was satisfied and no fallback was provided.
type NoBranchMatchedError struct{}

func (e *NoBranchMatchedError) Error() string {
	return "no branch condition was satisfied and no fallback was provided"
}

// Case represents a single case in a SwitchFlow.
type Case[T any, U any] struct {
	Predicate func(T) bool
	Execute   func() Result[U]
}

// SwitchFlow is a simpler alternative to BranchFlow for simple value-based switching.
// It takes a value and a slice of cases, plus a default.
func SwitchFlow[T, U any](
	value T,
	cases []Case[T, U],
	defaultCase func() Result[U],
) Result[U] {
	for _, c := range cases {
		if c.Predicate(value) {
			return c.Execute()
		}
	}

	return defaultCase()
}

// SwitchFlowWithResult is similar to SwitchFlow but the predicates operate on Result values.
func SwitchFlowWithResult[T, U any](
	result Result[T],
	cases []Case[T, U],
	defaultCase func() Result[U],
) Result[U] {
	if result.IsErr() {
		return Err[U](result.Error())
	}

	value := result.Value()

	for _, c := range cases {
		if c.Predicate(value) {
			return c.Execute()
		}
	}

	return defaultCase()
}
