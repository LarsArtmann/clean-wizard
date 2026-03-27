package context

import (
	"context"
	"maps"
	"slices"
)

// Context provides a generic, type-safe context holder that can be used across
// validation, error handling, and sanitization operations.
//
// This generic type supports multiple context use cases:
// - Context[ValidationConfig] for validation context
// - Context[ErrorConfig] for error context
// - Context[SanitizationConfig] for sanitization context
//
// Usage:
//
//	ctx := context.Background()
//	valCtx := context.NewContext[ValidationConfig](ctx, ValidationConfig{Field: "value"})
//	valCtx = valCtx.WithMetadata("trace_id", "123")
//
//	if val, ok := valCtx.GetMetadata("trace_id"); ok {
//	    // use val
//	}
type Context[T any] struct {
	// Base context (optional)
	Context context.Context //nolint:containedctx // Intentionally wraps context.Context for type-safe context passing

	// ValueType provides type-safe context data
	ValueType T

	// Metadata holds additional contextual information
	Metadata map[string]string

	// Permissions holds access control information
	Permissions []string
}

// NewContext creates a new Context with the provided value type.
func NewContext[T any](ctx context.Context, value T) *Context[T] {
	return &Context[T]{
		Context:     ctx,
		ValueType:   value,
		Metadata:    make(map[string]string),
		Permissions: []string{},
	}
}

// WithMetadata adds metadata to the context.
func (c *Context[T]) WithMetadata(key, value string) *Context[T] {
	c.Metadata[key] = value

	return c
}

// WithPermissions adds permissions to the context.
func (c *Context[T]) WithPermissions(permissions ...string) *Context[T] {
	c.Permissions = append(c.Permissions, permissions...)

	return c
}

// GetMetadata retrieves a metadata value by key.
func (c *Context[T]) GetMetadata(key string) (string, bool) {
	val, ok := c.Metadata[key]

	return val, ok
}

// HasPermission checks if a permission is present.
func (c *Context[T]) HasPermission(permission string) bool {
	return slices.Contains(c.Permissions, permission)
}

// Clone creates a deep copy of the context.
func (c *Context[T]) Clone() *Context[T] {
	metadata := make(map[string]string)
	maps.Copy(metadata, c.Metadata)

	permissions := slices.Clone(c.Permissions)

	return &Context[T]{
		Context:     c.Context,
		ValueType:   c.ValueType,
		Metadata:    metadata,
		Permissions: permissions,
	}
}

// Merge combines this context with another context of the same type.
// It creates a new context with merged metadata and permissions.
func (c *Context[T]) Merge(other *Context[T]) *Context[T] {
	cloned := c.Clone()

	// Merge metadata (other takes precedence on conflicts)
	maps.Copy(cloned.Metadata, other.Metadata)

	// Merge permissions (append)
	cloned.Permissions = append(cloned.Permissions, other.Permissions...)

	return cloned
}

// SetValueType updates the value type.
func (c *Context[T]) SetValueType(value T) *Context[T] {
	c.ValueType = value

	return c
}

// GetValueType returns the value type.
func (c *Context[T]) GetValueType() T {
	return c.ValueType
}

// GetContext returns the base context.Context, or context.Background() if nil.
func (c *Context[T]) GetContext() context.Context {
	if c.Context != nil {
		return c.Context
	}

	return context.Background()
}

// Branch creates a new context based on a condition.
// If condition is true, it applies the transform function to the current context.
// Otherwise, it returns the original context unchanged.
func (c *Context[T]) Branch(
	condition bool,
	transform func(*Context[T]) *Context[T],
) *Context[T] {
	if condition {
		return transform(c.Clone())
	}

	return c
}

// BranchWithValue creates a new context based on a predicate applied to the value.
// If predicate returns true, it applies the transform function.
func (c *Context[T]) BranchWithValue(
	predicate func(T) bool,
	transform func(*Context[T]) *Context[T],
) *Context[T] {
	if predicate(c.ValueType) {
		return transform(c.Clone())
	}

	return c.Clone()
}

// Fork creates multiple branched contexts based on conditions.
// Each pair in the map is a condition and the transform to apply.
// Returns a slice of contexts (only those where condition was true).
func (c *Context[T]) Fork(
	branches map[bool]func(*Context[T]) *Context[T],
) []*Context[T] {
	results := make([]*Context[T], 0, len(branches))

	for condition, transform := range branches {
		if condition {
			results = append(results, transform(c.Clone()))
		}
	}

	return results
}

// Join combines multiple contexts into one.
// The returned context has metadata and permissions merged from all inputs.
// For metadata conflicts, the last context's value wins.
// For permissions, all are appended.
func Join[T any](contexts ...*Context[T]) *Context[T] {
	if len(contexts) == 0 {
		return NewContext[T](context.Background(), *new(T))
	}

	if len(contexts) == 1 {
		return contexts[0].Clone()
	}

	result := contexts[0].Clone()

	for i := 1; i < len(contexts); i++ {
		result = result.Merge(contexts[i])
	}

	return result
}

// JoinWithValue combines multiple contexts and applies a value reducer.
// This is useful when you need to merge contexts and compute a new value.
func JoinWithValue[T, U any](
	contexts []*Context[T],
	valueReducer func([]T) U,
) *Context[U] {
	if len(contexts) == 0 {
		return NewContext[U](context.Background(), valueReducer([]T{}))
	}

	merged := contexts[0].Clone()
	values := make([]T, len(contexts))
	values[0] = contexts[0].ValueType

	for i := 1; i < len(contexts); i++ {
		merged = merged.Merge(contexts[i])
		values[i] = contexts[i].ValueType
	}

	return &Context[U]{
		Context:     merged.Context,
		ValueType:   valueReducer(values),
		Metadata:    merged.Metadata,
		Permissions: merged.Permissions,
	}
}

// Transform creates a new context with a transformed value type.
func Transform[T, U any](c *Context[T], fn func(T) U) *Context[U] {
	return &Context[U]{
		Context:     c.Context,
		ValueType:   fn(c.ValueType),
		Metadata:    c.Metadata,
		Permissions: c.Permissions,
	}
}

// WithContext sets a new base context.Context.
func (c *Context[T]) WithContext(ctx context.Context) *Context[T] {
	c.Context = ctx

	return c
}

// WithTimeout creates a new context with a timeout.
func (c *Context[T]) WithTimeout(timeout any) *Context[T] {
	// This is a simplified version - in practice you'd need the actual duration
	// For now, we just mark it as a placeholder
	c.Metadata["timeout_set"] = "true"

	return c
}

// Pick selects a context based on a predicate.
// Returns the first context where predicate returns true, or nil if none match.
func Pick[T any](contexts []*Context[T], predicate func(*Context[T]) bool) *Context[T] {
	for _, ctx := range contexts {
		if predicate(ctx) {
			return ctx
		}
	}

	return nil
}

// PickWithValue selects a context based on a predicate applied to its value.
// Returns the first context where predicate returns true, or nil if none match.
func PickWithValue[T any](contexts []*Context[T], predicate func(T) bool) *Context[T] {
	for _, ctx := range contexts {
		if predicate(ctx.ValueType) {
			return ctx
		}
	}

	return nil
}
