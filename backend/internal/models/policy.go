package models

import (
	"time"

	"github.com/google/uuid"
)

// Policy represents spending rules for an agent.
type Policy struct {
	ID                        uuid.UUID `json:"id"`
	AgentID                   uuid.UUID `json:"agent_id"`
	DailyLimitCents           int64     `json:"daily_limit_cents"`
	AllowedVendors            []string  `json:"allowed_vendors"`
	RequireApprovalAboveCents int64     `json:"require_approval_above_cents"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

// UpsertPolicyRequest is the request body for POST /policies.
type UpsertPolicyRequest struct {
	AgentID                   uuid.UUID `json:"agent_id"`
	DailyLimitCents           int64     `json:"daily_limit_cents"`
	AllowedVendors            []string  `json:"allowed_vendors"`
	RequireApprovalAboveCents int64     `json:"require_approval_above_cents"`
}
