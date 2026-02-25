package models

import (
	"time"

	"github.com/google/uuid"
)

// Transaction represents a spending transaction in the ledger.
type Transaction struct {
	ID          uuid.UUID              `json:"id"`
	RequestID   uuid.UUID              `json:"request_id"`
	AgentID     uuid.UUID              `json:"agent_id"`
	AmountCents int64                  `json:"amount_cents"`
	Currency    string                 `json:"currency"`
	Vendor      string                 `json:"vendor"`
	Status      string                 `json:"status"` // "approved", "denied", "pending_approval"
	Reason      string                 `json:"reason"`
	Meta        map[string]interface{} `json:"meta"`
	CreatedAt   time.Time              `json:"created_at"`
}

// SpendRequest is the request body for POST /spend.
type SpendRequest struct {
	RequestID uuid.UUID              `json:"request_id"`
	Amount    int64                  `json:"amount"` // cents
	Vendor    string                 `json:"vendor"`
	Meta      map[string]interface{} `json:"meta"`
}

// SpendResponse is the response for POST /spend.
type SpendResponse struct {
	Status string `json:"status"` // "approved" | "pending_approval" | "denied"
	Reason string `json:"reason"`
}

// ListTransactionsResponse is the response for GET /transactions.
type ListTransactionsResponse struct {
	Transactions []Transaction `json:"transactions"`
	PaginatedResponse
}
