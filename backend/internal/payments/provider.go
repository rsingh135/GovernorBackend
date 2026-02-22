package payments

import (
	"context"

	"github.com/google/uuid"
)

// Provider defines payment-provider integration for checkout and webhooks.
type Provider interface {
	Enabled() bool
	Name() string
	CreateCheckoutSession(ctx context.Context, req CreateCheckoutRequest) (*CheckoutSession, error)
	ParseWebhook(payload []byte, signature string) (*WebhookEvent, error)
}

// CreateCheckoutRequest is the provider-agnostic checkout request payload.
type CreateCheckoutRequest struct {
	TransactionID uuid.UUID
	AgentID       uuid.UUID
	AmountCents   int64
	Currency      string
	Vendor        string
}

// CheckoutSession is provider checkout metadata.
type CheckoutSession struct {
	Provider        string
	SessionID       string
	CheckoutURL     string
	PaymentIntentID string
	ProviderStatus  string
}

// WebhookEvent is normalized payment webhook payload.
type WebhookEvent struct {
	EventID         string
	Provider        string
	TransactionID   uuid.UUID
	ProviderStatus  string
	SessionID       string
	PaymentIntentID string
}
