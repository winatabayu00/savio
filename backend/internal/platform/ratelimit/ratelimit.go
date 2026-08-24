package ratelimit

import (
	"sync"
	"time"
)

// Limiter is an in-memory fixed-window rate limiter keyed by arbitrary string.
// Runnable check: one test. Use Redis for distributed deployments.
// ponytail: single-instance in-memory is enough for the take-home; swap to a
// Redis token bucket when Savio scales to multiple API replicas.
type Limiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	hits   map[string]*bucket
}

type bucket struct {
	count   int
	resetAt time.Time
}

func New(limit int, window time.Duration) *Limiter {
	return &Limiter{limit: limit, window: window, hits: make(map[string]*bucket)}
}

// Allow reports whether the key may proceed; it consumes one slot.
func (l *Limiter) Allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	b, ok := l.hits[key]
	if !ok || now.After(b.resetAt) {
		b = &bucket{count: 0, resetAt: now.Add(l.window)}
		l.hits[key] = b
	}
	b.count++
	return b.count <= l.limit
}
