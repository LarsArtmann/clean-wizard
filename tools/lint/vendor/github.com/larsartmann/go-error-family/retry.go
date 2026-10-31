package errorfamily

import "time"

// Default retry parameters for Transient-family errors.
const (
	defaultRetryMaxAttempts = 3
	defaultRetryMinDelay    = 100 * time.Millisecond
	defaultRetryMaxDelay    = 5 * time.Second
)

// RetryPolicy holds retry parameters derived from a Family. It is advisory: the
// library does not implement the retry loop — consumers (or retry libraries like
// failsafe-go) use these as sensible starting defaults.
//
// MaxAttempts is the total number of attempts including the first. Non-retryable
// families return MaxAttempts=1 (execute once, never retry).
type RetryPolicy struct {
	MaxAttempts int
	MinDelay    time.Duration
	MaxDelay    time.Duration
}

// RetryPolicy returns sensible retry defaults for this family.
// Only Transient is retryable; all other families return a single-attempt policy.
func (f Family) RetryPolicy() RetryPolicy {
	if f == Transient {
		return RetryPolicy{
			MaxAttempts: defaultRetryMaxAttempts,
			MinDelay:    defaultRetryMinDelay,
			MaxDelay:    defaultRetryMaxDelay,
		}
	}

	return RetryPolicy{MaxAttempts: 1}
}
