package models

import (
	"time"

	"github.com/google/uuid"
)

// Agent represents an authenticated AI agent linked to a user.
type Agent struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	Name         string    `json:"name"`
	Status       string    `json:"status"` // "active" or "frozen"
	APIKeyPrefix string    `json:"api_key_prefix"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateAgentRequest is the request body for POST /agents.
type CreateAgentRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Name   string    `json:"name"`
}

// CreateAgentResponse is the response body for POST /agents.
// api_key is only returned at creation time.
type CreateAgentResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
}
