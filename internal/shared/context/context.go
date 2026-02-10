package context

import (
	"context"
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
	Context context.Context

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
	for _, p := range c.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// Clone creates a deep copy of the context.
func (c *Context[T]) Clone() *Context[T] {
	metadata := make(map[string]string, len(c.Metadata))
	for k, v := range c.Metadata {
		metadata[k] = v
	}

	permissions := make([]string, len(c.Permissions))
	copy(permissions, c.Permissions)

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
	for k, v := range other.Metadata {
		cloned.Metadata[k] = v
	}

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