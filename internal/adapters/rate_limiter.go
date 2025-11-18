package adapters

import (
	"context"

	"golang.org/x/time/rate"
)

// RateLimiter provides rate limiting functionality for cleaning operations
type RateLimiter struct {
	limiter *rate.Limiter
}

// NewRateLimiter creates a new rate limiter
// rps: requests per second limit
// burst: maximum burst size
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(rps), burst),
	}
}

// Wait blocks until the operation is allowed or context is cancelled
func (rl *RateLimiter) Wait(ctx context.Context) error {
	return rl.limiter.Wait(ctx)
}

// Allow returns true if the operation is allowed immediately
func (rl *RateLimiter) Allow() bool {
	return rl.limiter.Allow()
}

// Reservation returns a reservation for a future operation
func (rl *RateLimiter) Reservation() *rate.Reservation {
	return rl.limiter.Reserve()
}

// RateLimitStats provides rate limiting statistics
type RateLimitStats struct {
	Limit   rate.Limit `json:"limit"`
	Burst   int        `json:"burst"`
	Tokens  float64    `json:"tokens"`
}

// Stats returns current rate limiter statistics
func (rl *RateLimiter) Stats() RateLimitStats {
	return RateLimitStats{
		Limit:  rl.limiter.Limit(),
		Burst:  rl.limiter.Burst(),
		Tokens: rl.limiter.Tokens(),
	}
}