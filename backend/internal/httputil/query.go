package httputil

import (
	"net/http"
	"strconv"
	"time"

	"agentpay/internal/models"

	"github.com/google/uuid"
)

// ParsePagination extracts pagination parameters from query string.
func ParsePagination(r *http.Request) models.PaginationParams {
	params := models.DefaultPagination()

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			params.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			params.Offset = offset
		}
	}

	params.Validate()
	return params
}

// ParseUUID extracts a UUID from query string by key.
func ParseUUID(r *http.Request, key string) (*uuid.UUID, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil, nil
	}

	id, err := uuid.Parse(value)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

// ParseTime extracts an ISO8601 timestamp from query string by key.
func ParseTime(r *http.Request, key string) (*time.Time, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil, nil
	}

	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// ParseString extracts a string from query string by key.
func ParseString(r *http.Request, key string) *string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil
	}
	return &value
}
