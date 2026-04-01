package middleware

import (
	"context"
	"fmt"
	"math"
	"net"
	"net/http"
	"sync"
	"time"
)

type bucket struct {
	mu         sync.Mutex
	tokens     float64
	lastRefill time.Time
}

// RateLimiter is a token bucket rate limiter backed by sync.Map.
type RateLimiter struct {
	buckets sync.Map
	rate    float64 // tokens per second
	burst   float64 // max tokens
}

// NewRateLimiter creates a new RateLimiter. Panics if rate <= 0.
func NewRateLimiter(ratePerSec, burst float64) *RateLimiter {
	if ratePerSec <= 0 {
		panic("ratelimit: ratePerSec must be > 0")
	}
	rl := &RateLimiter{rate: ratePerSec, burst: burst}
	go rl.cleanup(5 * time.Minute)
	return rl
}

// allow checks whether a request for the given key is allowed.
// Returns (allowed, remaining, resetAt).
func (rl *RateLimiter) allow(key string) (bool, int, time.Time) {
	now := time.Now()

	val, _ := rl.buckets.LoadOrStore(key, &bucket{
		tokens:     rl.burst,
		lastRefill: now,
	})
	b := val.(*bucket)

	b.mu.Lock()
	defer b.mu.Unlock()

	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens = math.Min(rl.burst, b.tokens+elapsed*rl.rate)
	b.lastRefill = now

	if b.tokens >= 1 {
		b.tokens--
		remaining := int(b.tokens)
		resetAt := now.Add(time.Duration((1-b.tokens)/rl.rate*float64(time.Second)))
		return true, remaining, resetAt
	}

	resetAt := now.Add(time.Duration((1-b.tokens)/rl.rate*float64(time.Second)))
	return false, 0, resetAt
}

func (rl *RateLimiter) writeHeaders(w http.ResponseWriter, remaining int, resetAt time.Time) {
	w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%.0f", rl.burst))
	w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
	w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetAt.Unix()))
}

// LimitByAPIKey rate-limits by the X-API-Key header value.
func (rl *RateLimiter) LimitByAPIKey(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key == "" {
			key = r.Header.Get("apiKey")
		}
		if key == "" {
			key = "anonymous"
		}
		allowed, remaining, resetAt := rl.allow("apikey:" + key)
		rl.writeHeaders(w, remaining, resetAt)
		if !allowed {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

// LimitByIP rate-limits by the client's IP address.
func (rl *RateLimiter) LimitByIP(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		allowed, remaining, resetAt := rl.allow("ip:" + ip)
		rl.writeHeaders(w, remaining, resetAt)
		if !allowed {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

// cleanup evicts buckets idle for more than 10 minutes, running every interval.
func (rl *RateLimiter) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-10 * time.Minute)
		rl.buckets.Range(func(key, val any) bool {
			b := val.(*bucket)
			b.mu.Lock()
			idle := b.lastRefill.Before(cutoff)
			b.mu.Unlock()
			if idle {
				rl.buckets.Delete(key)
			}
			return true
		})
	}
}

// cleanupWithContext is like cleanup but stops when ctx is cancelled.
// Exported so main.go can pass a cancellable context for clean shutdown.
func (rl *RateLimiter) CleanupWithContext(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			cutoff := time.Now().Add(-10 * time.Minute)
			rl.buckets.Range(func(key, val any) bool {
				b := val.(*bucket)
				b.mu.Lock()
				idle := b.lastRefill.Before(cutoff)
				b.mu.Unlock()
				if idle {
					rl.buckets.Delete(key)
				}
				return true
			})
		case <-ctx.Done():
			return
		}
	}
}
