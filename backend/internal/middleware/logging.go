package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware emits structured request logs.
type LoggingMiddleware struct{}

func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{}
}

func (m *LoggingMiddleware) Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		record := map[string]interface{}{
			"type":        "http_request",
			"request_id":  GetRequestID(r.Context()),
			"method":      r.Method,
			"path":        r.URL.Path,
			"status_code": rw.statusCode,
			"duration_ms": time.Since(start).Milliseconds(),
			"remote_addr": r.RemoteAddr,
			"at":          time.Now().UTC().Format(time.RFC3339Nano),
		}

		payload, err := json.Marshal(record)
		if err != nil {
			log.Printf("{\"type\":\"http_request\",\"error\":\"failed_to_marshal_log\"}")
			return
		}
		log.Println(string(payload))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
