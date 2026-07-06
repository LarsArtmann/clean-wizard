package execution

import (
	"context"
	"time"

	flow "github.com/Azure/go-workflow"
	"github.com/cenkalti/backoff/v4"
	errorfamily "github.com/larsartmann/go-error-family"
)

// RetryConfig controls per-step retry behavior for the workflow.
// When MaxAttempts is 0, retries are disabled.
type RetryConfig struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

// DefaultRetryConfig returns sensible defaults for cleaner retries:
// 3 attempts, starting at 2s, capped at 30s.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:    3,
		InitialBackoff: 2 * time.Second,
		MaxBackoff:     30 * time.Second,
	}
}

// RetryConfigFromAttempts returns a RetryConfig with the given max attempts
// and default backoff settings. This is the shared builder used by both
// clean and scan commands to avoid duplicating inline RetryConfig literals.
func RetryConfigFromAttempts(maxAttempts int) *RetryConfig {
	if maxAttempts <= 0 {
		return nil
	}
	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = maxAttempts
	return &cfg
}

// retryOptions converts a RetryConfig into go-workflow retry option funcs.
// The NextBackOff hook stops retrying immediately when the error is
// non-retryable (Infrastructure, Rejection, Conflict, Corruption) —
// retrying a permanent condition wastes time. Only Transient errors
// (timeouts, transient I/O, exec failures) proceed with backoff.
func retryOptions(cfg RetryConfig) []func(*flow.RetryOption) {
	if cfg.MaxAttempts <= 0 {
		return nil
	}

	initial := cfg.InitialBackoff
	if initial <= 0 {
		initial = 2 * time.Second
	}

	maxInterval := cfg.MaxBackoff
	if maxInterval <= 0 {
		maxInterval = 30 * time.Second
	}

	return []func(*flow.RetryOption){
		func(opt *flow.RetryOption) {
			opt.Attempts = uint64(cfg.MaxAttempts)

			expBackoff := backoff.NewExponentialBackOff(
				backoff.WithInitialInterval(initial),
				backoff.WithMaxInterval(maxInterval),
			)
			opt.Backoff = expBackoff

			opt.NextBackOff = func(_ context.Context, re flow.RetryEvent, _ time.Duration) time.Duration {
				if !errorfamily.IsRetryable(re.Error) {
					return backoff.Stop
				}

				return expBackoff.NextBackOff()
			}
		},
	}
}
