package models

import (
	"time"

	"github.com/google/uuid"
)

// ApprovalAuditLog records human review actions for pending transactions.
type ApprovalAuditLog struct {
	ID             uuid.UUID `json:"id"`
	TransactionID  uuid.UUID `json:"transaction_id"`
	AdminID        uuid.UUID `json:"admin_id"`
	AdminEmail     string    `json:"admin_email"`
	Action         string    `json:"action"`
	PreviousStatus string    `json:"previous_status"`
	NewStatus      string    `json:"new_status"`
	Reason         string    `json:"reason"`
	RequestID      string    `json:"request_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// CreateApprovalAuditLogRequest is the write model for repository inserts.
type CreateApprovalAuditLogRequest struct {
	TransactionID  uuid.UUID
	AdminID        uuid.UUID
	Action         string
	PreviousStatus string
	NewStatus      string
	Reason         string
	RequestID      string
}
