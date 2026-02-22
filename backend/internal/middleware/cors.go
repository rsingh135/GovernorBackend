package middleware

import (
	"net/http"
	"strings"
)

// CORSMiddleware handles browser preflight and origin allowlisting.
type CORSMiddleware struct {
	allowedOrigins map[string]struct{}
	allowAll       bool
}

// NewCORSMiddleware creates a CORS middleware with exact-match origin allowlisting.
func NewCORSMiddleware(origins []string) *CORSMiddleware {
	allowed := make(map[string]struct{}, len(origins))
	allowAll := false

	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		if trimmed == "*" {
			allowAll = true
			continue
		}
		allowed[trimmed] = struct{}{}
	}

	return &CORSMiddleware{
		allowedOrigins: allowed,
		allowAll:       allowAll,
	}
}

// Handle wraps an http.Handler and applies CORS headers plus preflight responses.
func (m *CORSMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		originAllowed := origin != "" && (m.allowAll || m.isAllowed(origin))

		if originAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, apiKey, X-Request-Id")
			w.Header().Set("Access-Control-Max-Age", "600")
			appendVaryHeader(w, "Origin")
		}

		if r.Method == http.MethodOptions {
			if origin == "" || originAllowed {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			http.Error(w, "origin not allowed", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *CORSMiddleware) isAllowed(origin string) bool {
	_, ok := m.allowedOrigins[origin]
	return ok
}

func appendVaryHeader(w http.ResponseWriter, value string) {
	current := strings.TrimSpace(w.Header().Get("Vary"))
	if current == "" {
		w.Header().Set("Vary", value)
		return
	}

	parts := strings.Split(current, ",")
	for _, part := range parts {
		if strings.EqualFold(strings.TrimSpace(part), value) {
			return
		}
	}
	w.Header().Set("Vary", current+", "+value)
}
