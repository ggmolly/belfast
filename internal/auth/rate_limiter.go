package auth

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateEntry
}

type rateEntry struct {
	count   int
	resetAt time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{entries: make(map[string]*rateEntry)}
}

func (limiter *RateLimiter) Allow(key string, limit int, window time.Duration) bool {
	if limit <= 0 {
		return true
	}
	if window <= 0 {
		return true
	}
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	now := time.Now()
	entry, ok := limiter.entries[key]
	if !ok || now.After(entry.resetAt) {
		limiter.entries[key] = &rateEntry{count: 1, resetAt: now.Add(window)}
		return true
	}
	if entry.count >= limit {
		return false
	}
	entry.count++
	return true
}
