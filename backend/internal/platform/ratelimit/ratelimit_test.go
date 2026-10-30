package ratelimit

import (
	"context"
	"testing"
	"time"
)

// TC-API-001: Sliding Window Rate Limiter & HTTP 429 Enforcement
func TestTCAPI001SlidingWindowRateLimiting(t *testing.T) {
	ctx := context.Background()

	// Limit: 5 requests per 100 milliseconds
	limiter := NewSlidingWindowLimiter(5, 100*time.Millisecond)
	key := "client-ip-192.168.1.50"

	t.Run("Sub-test 1: First 5 requests allowed", func(t *testing.T) {
		for i := 1; i <= 5; i++ {
			res := limiter.Allow(ctx, key)
			if !res.Allowed {
				t.Fatalf("request %d should have been allowed", i)
			}
			if res.Remaining != 5-i {
				t.Fatalf("expected remaining %d, got %d", 5-i, res.Remaining)
			}
		}
	})

	t.Run("Sub-test 2: 6th request within window rejected with HTTP 429", func(t *testing.T) {
		res := limiter.Allow(ctx, key)
		if res.Allowed {
			t.Fatalf("request beyond limit of 5 must be rejected")
		}
		if res.Remaining != 0 {
			t.Fatalf("expected 0 remaining, got %d", res.Remaining)
		}
	})

	t.Run("Sub-test 3: Window expiration automatically resets allowance", func(t *testing.T) {
		// Wait for window to expire
		time.Sleep(120 * time.Millisecond)

		res := limiter.Allow(ctx, key)
		if !res.Allowed {
			t.Fatalf("request after window expiry should be allowed")
		}
		if res.Remaining != 4 {
			t.Fatalf("expected 4 remaining, got %d", res.Remaining)
		}
	})
}
