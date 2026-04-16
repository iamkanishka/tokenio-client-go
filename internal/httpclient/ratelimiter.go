package httpclient

import (
	"context"
	"sync"
	"time"
)

// RateLimiter is a simple token-bucket rate limiter using only stdlib.
// It limits the rate of HTTP requests to protect against API rate limit errors.
type RateLimiter struct {
	mu       sync.Mutex
	tokens   float64
	maxBurst float64
	ratePerS float64 // tokens added per second
	lastTick time.Time
}

// NewRateLimiter creates a RateLimiter that allows rps requests per second
// with a burst capacity of burst.
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		tokens:   float64(burst),
		maxBurst: float64(burst),
		ratePerS: rps,
		lastTick: time.Now(),
	}
}

// Wait blocks until a token is available or ctx is cancelled.
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		r.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(r.lastTick).Seconds()
		r.tokens += elapsed * r.ratePerS
		if r.tokens > r.maxBurst {
			r.tokens = r.maxBurst
		}
		r.lastTick = now

		if r.tokens >= 1 {
			r.tokens--
			r.mu.Unlock()
			return nil
		}

		// Calculate wait time for next token.
		needed := 1 - r.tokens
		waitDur := time.Duration(needed/r.ratePerS*1000) * time.Millisecond
		r.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitDur):
		}
	}
}
