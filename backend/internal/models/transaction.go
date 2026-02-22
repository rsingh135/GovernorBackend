package models

import (
	"time"

	"github.com/google/uuid"
)

// Transaction represents a spending transaction in the ledger.
type Transaction struct {
	ID                      uuid.UUID              `json:"id"`
	RequestID               uuid.UUID              `json:"request_id"`
	AgentID                 uuid.UUID              `json:"agent_id"`
	AmountCents             int64                  `json:"amount_cents"`
	Currency                string                 `json:"currency"`
	Vendor                  string                 `json:"vendor"`
	Status                  string                 `json:"status"` // "approved", "denied", "pending_approval"
	Reason                  string                 `json:"reason"`
	Provider                string                 `json:"provider,omitempty"`
	ProviderSessionID       string                 `json:"provider_session_id,omitempty"`
	ProviderPaymentIntentID string                 `json:"provider_payment_intent_id,omitempty"`
	ProviderStatus          string                 `json:"provider_status,omitempty"`
	ProviderCheckoutURL     string                 `json:"provider_checkout_url,omitempty"`
	Meta                    map[string]interface{} `json:"meta"`
	CreatedAt               time.Time              `json:"created_at"`
}

// SpendRequest is the request body for POST /spend.
type SpendRequest struct {
	RequestID uuid.UUID              `json:"request_id"`
	Amount    int64                  `json:"amount"` // cents
	Vendor    string                 `json:"vendor"`
	MCC       string                 `json:"mcc,omitempty"`
	Meta      map[string]interface{} `json:"meta"`
}

// SpendResponse is the response for POST /spend.
type SpendResponse struct {
	Status         string `json:"status"` // "approved" | "pending_approval" | "denied"
	Reason         string `json:"reason"`
	CheckoutURL    string `json:"checkout_url,omitempty"`
	ProviderStatus string `json:"provider_status,omitempty"`
	TransactionID  string `json:"transaction_id,omitempty"`
}
