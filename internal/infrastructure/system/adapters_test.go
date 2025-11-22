package system_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRateLimiter(t *testing.T) {
	rl := adapters.NewRateLimiter(10, 5)
	require.NotNil(t, rl)

	stats := rl.Stats()
	assert.Greater(t, stats.Limit, float64(0))
	assert.Equal(t, 5, stats.Burst)
	assert.GreaterOrEqual(t, stats.Tokens, float64(0))
}

func TestRateLimiter_Allow(t *testing.T) {
	rl := adapters.NewRateLimiter(100, 10) // High limit for testing
	require.NotNil(t, rl)

	// Should allow immediate request
	assert.True(t, rl.Allow())

	// With high limit, should still allow
	assert.True(t, rl.Allow())
}

func TestRateLimiter_Wait(t *testing.T) {
	rl := adapters.NewRateLimiter(100, 10) // High limit for testing
	require.NotNil(t, rl)

	ctx := context.Background()

	// Should not error with high limit
	err := rl.Wait(ctx)
	assert.NoError(t, err)
}

func TestCacheManager(t *testing.T) {
	cm := adapters.NewCacheManager(5*time.Minute, 10*time.Minute)
	require.NotNil(t, cm)

	// Test Set and Get
	cm.Set("test_key", "test_value", time.Minute)

	value, found := cm.Get("test_key")
	assert.True(t, found)
	assert.Equal(t, "test_value", value)

	// Test Get for non-existent key
	_, found = cm.Get("non_existent")
	assert.False(t, found)

	// Test ItemCount
	assert.Equal(t, 1, cm.ItemCount())

	// Test Delete
	cm.Delete("test_key")
	_, found = cm.Get("test_key")
	assert.False(t, found)
	assert.Equal(t, 0, cm.ItemCount())
}

func TestCacheManager_Expiration(t *testing.T) {
	cm := adapters.NewCacheManager(100*time.Millisecond, 1*time.Minute)
	require.NotNil(t, cm)

	cm.Set("expire_key", "expire_value", 50*time.Millisecond)

	// Should exist immediately
	value, found := cm.Get("expire_key")
	assert.True(t, found)
	assert.Equal(t, "expire_value", value)

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Should be expired
	_, found = cm.Get("expire_key")
	assert.False(t, found)
}

func TestCacheManager_Clear(t *testing.T) {
	cm := adapters.NewCacheManager(5*time.Minute, 10*time.Minute)
	require.NotNil(t, cm)

	// Add multiple items
	cm.Set("key1", "value1", time.Minute)
	cm.Set("key2", "value2", time.Minute)
	cm.Set("key3", "value3", time.Minute)

	assert.Equal(t, 3, cm.ItemCount())

	// Clear all
	cm.Clear()
	assert.Equal(t, 0, cm.ItemCount())
}

func TestHTTPClient(t *testing.T) {
	client := adapters.NewHTTPClient()
	require.NotNil(t, client)

	// Test WithTimeout
	timeoutClient := client.WithTimeout(5 * time.Second)
	require.NotNil(t, timeoutClient)

	// Test WithRetry
	retryClient := client.WithRetry(5, 100*time.Millisecond, 2*time.Second)
	require.NotNil(t, retryClient)

	// Test WithAuth
	authClient := client.WithAuth("Bearer", "test-token")
	require.NotNil(t, authClient)

	// Test WithHeader
	headerClient := client.WithHeader("X-Custom", "test-value")
	require.NotNil(t, headerClient)
}

func TestEnvironmentConfig(t *testing.T) {
	cfg, err := adapters.LoadEnvironmentConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Test default values
	assert.False(t, cfg.Debug)
	assert.Equal(t, "development", cfg.Environment)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, 4, cfg.MaxConcurrency)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.Equal(t, 10.0, cfg.RateLimitRPS)
	assert.True(t, cfg.CacheEnabled)
	assert.Equal(t, 5*time.Minute, cfg.CacheTTL)
}

func TestEnvironmentConfig_Validate(t *testing.T) {
	cfg, err := adapters.LoadEnvironmentConfig()
	require.NoError(t, err)

	// Valid config should pass
	err = cfg.ValidateEnvironmentConfig()
	assert.NoError(t, err)
}

func TestEnvironmentConfig_Getters(t *testing.T) {
	// Test GetEnvWithDefault
	value := adapters.GetEnvWithDefault("NON_EXISTENT_ENV", "default")
	assert.Equal(t, "default", value)

	// Test GetEnvBool
	isTrue := adapters.GetEnvBool("DEBUG", false)
	assert.False(t, isTrue) // Should be default since DEBUG is not set

	// Test GetEnvInt
	num := adapters.GetEnvInt("NON_EXISTENT_INT", 42)
	assert.Equal(t, 42, num)

	// Test GetEnvDuration
	dur := adapters.GetEnvDuration("NON_EXISTENT_DURATION", 30*time.Second)
	assert.Equal(t, 30*time.Second, dur)
}

func TestErrorConstructors(t *testing.T) {
	// Test ErrInvalidConfig
	err := adapters.ErrInvalidConfig("test error")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration error: test error")

	// Test ErrInvalidArgument
	err = adapters.ErrInvalidArgument("arg", "test message")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid argument 'arg': test message")

	// Test ErrNotFound
	err = adapters.ErrNotFound("resource")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resource not found: resource")

	// Test ErrTimeout
	err = adapters.ErrTimeout("operation")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation timeout: operation")

	// Test ErrRateLimit
	err = adapters.ErrRateLimit(10.5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit exceeded: 10.50 requests/second")

	// Test ErrCacheMiss
	err = adapters.ErrCacheMiss("key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cache miss: key 'key' not found")

	// Test ErrNotImplemented
	err = adapters.ErrNotImplemented("feature")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "feature not implemented: feature")

	// Test ErrUnauthorized
	err = adapters.ErrUnauthorized("action")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized: action")

	// Test ErrForbidden
	err = adapters.ErrForbidden("action")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden: action")

	// Test ErrServiceUnavailable
	err = adapters.ErrServiceUnavailable("service")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service unavailable: service")
}
