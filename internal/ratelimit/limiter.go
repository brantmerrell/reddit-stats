package ratelimit

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

type RateLimiter struct {
	mu        sync.Mutex
	remaining int
	reset     time.Time
	limit     int
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		remaining: 600,
		limit:     600,
		reset:     time.Now().Add(10 * time.Minute),
	}
}

func (r *RateLimiter) Wait() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.remaining <= 0 {
		waitTime := time.Until(r.reset)
		if waitTime > 0 {
			time.Sleep(waitTime)
		}
		r.remaining = r.limit
		r.reset = time.Now().Add(10 * time.Minute)
	}

	r.remaining--
	return nil
}

func (r *RateLimiter) UpdateFromHeaders(headers http.Header) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if remaining := headers.Get("X-Ratelimit-Remaining"); remaining != "" {
		if val, err := strconv.Atoi(remaining); err == nil {
			r.remaining = val
		}
	}

	if reset := headers.Get("X-Ratelimit-Reset"); reset != "" {
		if val, err := strconv.ParseInt(reset, 10, 64); err == nil {
			r.reset = time.Unix(val, 0)
		}
	}

	if limit := headers.Get("X-Ratelimit-Limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			r.limit = val
		}
	}
}
