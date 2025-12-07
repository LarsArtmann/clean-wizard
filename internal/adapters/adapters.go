package adapters

import (
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/infrastructure/system"
)

// Re-export system package functions for compatibility

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rps float64, burst int) *system.RateLimiter {
	return system.NewRateLimiter(rps, burst)
}

// NewCacheManager creates a new cache manager
func NewCacheManager(defaultExpiration, cleanupInterval time.Duration) *system.CacheManager {
	return system.NewCacheManager(defaultExpiration, cleanupInterval)
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient() *system.HTTPClient {
	return system.NewHTTPClient()
}

// LoadEnvironmentConfig loads configuration from environment variables
func LoadEnvironmentConfig() (*system.EnvironmentConfig, error) {
	return system.LoadEnvironmentConfig()
}

// GetEnvWithDefault returns environment variable with default value
func GetEnvWithDefault(key, defaultValue string) string {
	return system.GetEnvWithDefault(key, defaultValue)
}

// GetEnvBool returns boolean environment variable with default
func GetEnvBool(key string, defaultValue bool) bool {
	return system.GetEnvBool(key, defaultValue)
}

// GetEnvInt returns integer environment variable with default
func GetEnvInt(key string, defaultValue int) int {
	return system.GetEnvInt(key, defaultValue)
}

// GetEnvDuration returns duration environment variable with default
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	return system.GetEnvDuration(key, defaultValue)
}

// Error constructors
func ErrInvalidConfig(message string, args ...any) error {
	return system.ErrInvalidConfig(message, args...)
}

func ErrInvalidArgument(arg, message string) error {
	return system.ErrInvalidArgument(arg, message)
}

func ErrNotFound(resource string) error {
	return system.ErrNotFound(resource)
}

func ErrTimeout(operation string) error {
	return system.ErrTimeout(operation)
}

func ErrRateLimit(limit float64) error {
	return system.ErrRateLimit(limit)
}

func ErrCacheMiss(key string) error {
	return system.ErrCacheMiss(key)
}

func ErrNotImplemented(feature string) error {
	return system.ErrNotImplemented(feature)
}

func ErrUnauthorized(operation string) error {
	return system.ErrUnauthorized(operation)
}

func ErrForbidden(operation string) error {
	return system.ErrForbidden(operation)
}

func ErrServiceUnavailable(service string) error {
	return system.ErrServiceUnavailable(service)
}