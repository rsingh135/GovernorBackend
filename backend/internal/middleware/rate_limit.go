package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiterMiddleware provides a simple in-memory fixed-window limiter.
type RateLimiterMiddleware struct {
	limit  int
	window time.Duration
	mu     sync.Mutex
	state  map[string]*rateLimitWindow
}

type rateLimitWindow struct {
	start time.Time
	count int
}

func NewRateLimiterMiddleware(limit int, window time.Duration) *RateLimiterMiddleware {
	if limit <= 0 {
		limit = 60
	}
	if window <= 0 {
		window = time.Minute
	}

	return &RateLimiterMiddleware{
		limit:  limit,
		window: window,
		state:  make(map[string]*rateLimitWindow),
	}
}

func (m *RateLimiterMiddleware) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := clientKey(r)
		now := time.Now()

		m.mu.Lock()
		window, ok := m.state[key]
		if !ok || now.Sub(window.start) >= m.window {
			window = &rateLimitWindow{start: now, count: 0}
			m.state[key] = window
		}

		if window.count >= m.limit {
			retryAfter := int(m.window.Seconds() - now.Sub(window.start).Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			m.mu.Unlock()

			w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		window.count++
		m.mu.Unlock()

		next.ServeHTTP(w, r)
	}
}

func clientKey(r *http.Request) string {
	ip := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
