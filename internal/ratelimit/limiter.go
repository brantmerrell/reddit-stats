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
    buffer        float64
    windowSize    time.Duration
    requestTimes  []time.Time
}

func NewRateLimiter() *RateLimiter {
    return &RateLimiter{
        remaining: 600,
        limit:     600,
        reset:     time.Now().Add(10 * time.Minute),
        buffer:      0.1,
        windowSize:  10 * time.Minute,
        requestTimes: make([]time.Time, 0, 600),
    }
}

func (r *RateLimiter) Wait() error {
    r.mu.Lock()
    defer r.mu.Unlock()

    now := time.Now()
    r.cleanOldRequests(now)

    windowDuration := r.windowSize
    idealSpacing := windowDuration / time.Duration(r.limit)

    if len(r.requestTimes) > 0 {
        lastRequest := r.requestTimes[len(r.requestTimes)-1]
        timeSinceLastRequest := now.Sub(lastRequest)
        
        if timeSinceLastRequest < idealSpacing {
            sleepTime := idealSpacing - timeSinceLastRequest
            r.mu.Unlock()
            time.Sleep(sleepTime)
            r.mu.Lock()
            now = time.Now()
        }
    }

    bufferLimit := int(float64(r.limit) * (1 - r.buffer))
    if len(r.requestTimes) >= bufferLimit {
        timeToReset := time.Until(r.reset)
        if timeToReset > 0 {
            r.mu.Unlock()
            time.Sleep(timeToReset)
            r.mu.Lock()
            r.requestTimes = make([]time.Time, 0, r.limit)
            r.reset = time.Now().Add(r.windowSize)
        }
    }

    r.requestTimes = append(r.requestTimes, now)
    r.remaining = r.limit - len(r.requestTimes)

    return nil
}

func (r *RateLimiter) cleanOldRequests(now time.Time) {
    cutoff := now.Add(-r.windowSize)
    newIdx := 0
    
    for i, t := range r.requestTimes {
        if t.After(cutoff) {
            newIdx = i
            break
        }
    }
    
    if newIdx > 0 {
        r.requestTimes = r.requestTimes[newIdx:]
    }
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

