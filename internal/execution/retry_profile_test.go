package execution

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetryProfile_IsValid(t *testing.T) {
	t.Parallel()

	valid := []RetryProfile{
		RetryProfileDefault, RetryProfileAggressive,
		RetryProfileConservative, RetryProfileNone,
	}
	for _, p := range valid {
		assert.True(t, p.IsValid(), "%q should be valid", p)
	}

	invalid := []RetryProfile{"", "fast", "slow", "unknown"}
	for _, p := range invalid {
		if p == "" {
			continue // empty string is treated as Default by Apply, but IsValid returns false
		}
		assert.False(t, p.IsValid(), "%q should be invalid", p)
	}
}

func TestRetryProfile_Apply(t *testing.T) {
	t.Parallel()

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		cfg := RetryProfileDefault.Apply()
		assert.NotNil(t, cfg)
		assert.Equal(t, 3, cfg.MaxAttempts)
		assert.Equal(t, 2*time.Second, cfg.InitialBackoff)
		assert.Equal(t, 30*time.Second, cfg.MaxBackoff)
	})

	t.Run("aggressive", func(t *testing.T) {
		t.Parallel()

		cfg := RetryProfileAggressive.Apply()
		assert.NotNil(t, cfg)
		assert.Equal(t, 5, cfg.MaxAttempts)
		assert.Equal(t, 1*time.Second, cfg.InitialBackoff)
		assert.Equal(t, 60*time.Second, cfg.MaxBackoff)
	})

	t.Run("conservative", func(t *testing.T) {
		t.Parallel()

		cfg := RetryProfileConservative.Apply()
		assert.NotNil(t, cfg)
		assert.Equal(t, 2, cfg.MaxAttempts)
		assert.Equal(t, 5*time.Second, cfg.InitialBackoff)
		assert.Equal(t, 30*time.Second, cfg.MaxBackoff)
	})

	t.Run("none disables retries", func(t *testing.T) {
		t.Parallel()

		cfg := RetryProfileNone.Apply()
		assert.Nil(t, cfg)
	})

	t.Run("empty defaults to Default", func(t *testing.T) {
		t.Parallel()

		cfg := RetryProfile("").Apply()
		assert.NotNil(t, cfg)
		assert.Equal(t, 3, cfg.MaxAttempts)
	})
}
