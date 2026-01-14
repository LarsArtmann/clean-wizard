package adapters

import (
	"fmt"
)

// ErrInvalidConfig creates a configuration validation error.
func ErrInvalidConfig(message string) error {
	return fmt.Errorf("configuration error: %s", message)
}

// ErrInvalidArgument creates an argument validation error.
func ErrInvalidArgument(arg, message string) error {
	return fmt.Errorf("invalid argument '%s': %s", arg, message)
}

// ErrNotFound creates a not found error.
func ErrNotFound(resource string) error {
	return fmt.Errorf("resource not found: %s", resource)
}

// ErrTimeout creates a timeout error.
func ErrTimeout(operation string) error {
	return fmt.Errorf("operation timeout: %s", operation)
}

// ErrRateLimit creates a rate limit error.
func ErrRateLimit(limit float64) error {
	return fmt.Errorf("rate limit exceeded: %.2f requests/second", limit)
}

// ErrCacheMiss creates a cache miss error.
func ErrCacheMiss(key string) error {
	return fmt.Errorf("cache miss: key '%s' not found", key)
}

// ErrHTTPError creates an HTTP error.
func ErrHTTPError(statusCode int, message string) error {
	return fmt.Errorf("HTTP error %d: %s", statusCode, message)
}

// ErrNotImplemented creates a not implemented error.
func ErrNotImplemented(feature string) error {
	return fmt.Errorf("feature not implemented: %s", feature)
}

// ErrUnauthorized creates an unauthorized error.
func ErrUnauthorized(operation string) error {
	return fmt.Errorf("unauthorized: %s", operation)
}

// ErrForbidden creates a forbidden error.
func ErrForbidden(operation string) error {
	return fmt.Errorf("forbidden: %s", operation)
}

// ErrServiceUnavailable creates a service unavailable error.
func ErrServiceUnavailable(service string) error {
	return fmt.Errorf("service unavailable: %s", service)
}
