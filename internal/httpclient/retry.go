package httpclient

import (
	"crypto/rand"
	"encoding/binary"
	"math"
	"time"
)

// retryPolicy defines exponential backoff with full jitter.
type retryPolicy struct {
	maxRetries int
	waitMin    time.Duration
	waitMax    time.Duration
}

// backoff returns the wait duration for attempt n (1-indexed).
// Uses crypto/rand for jitter so gosec G404 is not triggered.
func (r *retryPolicy) backoff(n int) time.Duration {
	cap := float64(r.waitMax)
	base := float64(r.waitMin) * math.Pow(2, float64(n))
	if base > cap {
		base = cap
	}
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return r.waitMin // extremely unlikely; fall back to minimum
	}
	rf := float64(binary.BigEndian.Uint64(b[:])) / float64(math.MaxUint64)
	jittered := time.Duration(rf * base)
	if jittered < r.waitMin {
		return r.waitMin
	}
	return jittered
}
