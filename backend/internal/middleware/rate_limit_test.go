package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterMiddleware_LimitsByClient(t *testing.T) {
	mw := NewRateLimiterMiddleware(1, time.Minute)
	h := mw.Limit(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	req1 := httptest.NewRequest(http.MethodPost, "/admin/login", nil)
	req1.RemoteAddr = "127.0.0.1:12345"
	rr1 := httptest.NewRecorder()
	h(rr1, req1)
	if rr1.Code != http.StatusNoContent {
		t.Fatalf("expected first request status 204, got %d", rr1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/admin/login", nil)
	req2.RemoteAddr = "127.0.0.1:12346"
	rr2 := httptest.NewRecorder()
	h(rr2, req2)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request status 429, got %d", rr2.Code)
	}
}
