package auth

import (
	"sync"
	"time"
)

type bucket struct {
	count int
	reset time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	limit   int
	window  time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*bucket),
		limit:   limit,
		window:  window,
	}
}

func (r *RateLimiter) Allow(key string) (bool, time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	if len(r.buckets) > 4096 { // evita crescimento sem limite com chaves variadas
		for k, v := range r.buckets {
			if now.After(v.reset) {
				delete(r.buckets, k)
			}
		}
	}
	b, ok := r.buckets[key]
	if !ok || now.After(b.reset) {
		r.buckets[key] = &bucket{count: 1, reset: now.Add(r.window)}
		return true, 0
	}
	if b.count >= r.limit {
		return false, b.reset.Sub(now)
	}
	b.count++
	return true, 0
}
