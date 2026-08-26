package api

import (
	"sync"
	"time"
)

// RateLimiter is a small in-memory fixed-window limiter for local and FC runtime guards.
type RateLimiter struct {
	mu     sync.Mutex
	max    int
	window time.Duration
	hits   map[string][]time.Time
}

// NewRateLimiter creates a request limiter. A non-positive max disables limiting.
func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		max:    max,
		window: window,
		hits:   make(map[string][]time.Time),
	}
}

func (l *RateLimiter) allow(key string, now time.Time) (bool, time.Duration) {
	if l == nil || l.max <= 0 || l.window <= 0 {
		return true, 0
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := now.Add(-l.window)
	requests := l.hits[key]
	keepFrom := 0
	for keepFrom < len(requests) && requests[keepFrom].Before(cutoff) {
		keepFrom++
	}
	requests = append(requests[:0], requests[keepFrom:]...)

	if len(requests) >= l.max {
		retryAfter := l.window
		if len(requests) > 0 {
			retryAfter = requests[0].Add(l.window).Sub(now)
			if retryAfter < time.Second {
				retryAfter = time.Second
			}
		}
		l.hits[key] = requests
		return false, retryAfter
	}

	requests = append(requests, now)
	l.hits[key] = requests
	return true, 0
}
