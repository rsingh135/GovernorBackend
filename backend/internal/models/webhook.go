package models

import (
	"time"

	"github.com/google/uuid"
)

// Webhook is a registered endpoint for event notifications.
type Webhook struct {
	ID        uuid.UUID `json:"id"`
	AgentID   uuid.UUID `json:"agent_id"`
	URL       string    `json:"url"`
	// Secret intentionally omitted from JSON — never returned to callers.
	Events    []string  `json:"events"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// WebhookWithSecret extends Webhook with the signing secret (internal use only).
type WebhookWithSecret struct {
	Webhook
	Secret string
}

// WebhookDelivery tracks a single delivery attempt.
type WebhookDelivery struct {
	ID             uuid.UUID  `json:"id"`
	WebhookID      uuid.UUID  `json:"webhook_id"`
	TransactionID  uuid.UUID  `json:"transaction_id"`
	Event          string     `json:"event"`
	Payload        []byte     `json:"-"`
	Status         string     `json:"status"`
	Attempt        int        `json:"attempt"`
	NextAttemptAt  time.Time  `json:"next_attempt_at"`
	LastAttemptAt  *time.Time `json:"last_attempt_at,omitempty"`
	LastHTTPStatus *int       `json:"last_http_status,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// RegisterWebhookRequest is the request body for POST /webhooks.
type RegisterWebhookRequest struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
	Secret string   `json:"secret"`
}

// RegisterWebhookResponse is the response for POST /webhooks.
type RegisterWebhookResponse struct {
	ID        uuid.UUID `json:"id"`
	URL       string    `json:"url"`
	Events    []string  `json:"events"`
	CreatedAt time.Time `json:"created_at"`
}

// ListWebhooksResponse is the response for GET /webhooks.
type ListWebhooksResponse struct {
	Webhooks []Webhook `json:"webhooks"`
}

// WebhookEventPayload is the signed body sent to registered endpoints.
type WebhookEventPayload struct {
	EventID       uuid.UUID `json:"event_id"`
	Event         string    `json:"event"`
	TransactionID uuid.UUID `json:"transaction_id"`
	AgentID       uuid.UUID `json:"agent_id"`
	AmountCents   int64     `json:"amount_cents"`
	Vendor        string    `json:"vendor"`
	Status        string    `json:"status"`
	Timestamp     time.Time `json:"timestamp"`
}
