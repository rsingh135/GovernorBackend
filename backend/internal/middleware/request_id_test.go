package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDMiddleware_GeneratesAndPropagates(t *testing.T) {
	mw := NewRequestIDMiddleware()

	h := mw.AddRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if GetRequestID(r.Context()) == "" {
			t.Fatalf("expected request id in context")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Header().Get("X-Request-Id") == "" {
		t.Fatalf("expected X-Request-Id response header")
	}
}

func TestRequestIDMiddleware_UsesIncomingHeader(t *testing.T) {
	mw := NewRequestIDMiddleware()
	const incomingID = "req_test_custom"

	h := mw.AddRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := GetRequestID(r.Context()); got != incomingID {
			t.Fatalf("expected request id %s, got %s", incomingID, got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Request-Id", incomingID)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if got := rr.Header().Get("X-Request-Id"); got != incomingID {
		t.Fatalf("expected response request id %s, got %s", incomingID, got)
	}
}
