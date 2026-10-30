package ratelimit

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrRateLimitExceeded = errors.New("HTTP_429_TOO_MANY_REQUESTS")
)

type RateLimitResult struct {
	Allowed   bool
	Remaining int
	ResetIn   time.Duration
}

type SlidingWindowLimiter struct {
	mu      sync.Mutex
	windows map[string][]time.Time // key -> slice of request timestamps
	window  time.Duration
	limit   int
}

func NewSlidingWindowLimiter(limit int, window time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		windows: make(map[string][]time.Time),
		window:  window,
		limit:   limit,
	}
}

// Allow checks if request under key is allowed within sliding window
func (l *SlidingWindowLimiter) Allow(ctx context.Context, key string) RateLimitResult {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().UTC()
	cutoff := now.Add(-l.window)

	// Clean up timestamps older than window
	existing := l.windows[key]
	valid := make([]time.Time, 0, len(existing))
	for _, t := range existing {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= l.limit {
		oldest := valid[0]
		resetIn := oldest.Add(l.window).Sub(now)
		if resetIn < 0 {
			resetIn = 0
		}
		l.windows[key] = valid
		return RateLimitResult{
			Allowed:   false,
			Remaining: 0,
			ResetIn:   resetIn,
		}
	}

	valid = append(valid, now)
	l.windows[key] = valid

	return RateLimitResult{
		Allowed:   true,
		Remaining: l.limit - len(valid),
		ResetIn:   l.window,
	}
}
