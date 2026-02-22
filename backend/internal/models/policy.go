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
	PerTransactionLimitCents  int64     `json:"per_transaction_limit_cents"`
	AllowedVendors            []string  `json:"allowed_vendors"`
	AllowedMCCs               []string  `json:"allowed_mccs"`
	AllowedWeekdaysUTC        []int     `json:"allowed_weekdays_utc"`
	AllowedHoursUTC           []int     `json:"allowed_hours_utc"`
	RequireApprovalAboveCents int64     `json:"require_approval_above_cents"`
	PurchaseGuideline         string    `json:"purchase_guideline"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

// UpsertPolicyRequest is the request body for POST /policies.
type UpsertPolicyRequest struct {
	AgentID                   uuid.UUID `json:"agent_id"`
	DailyLimitCents           int64     `json:"daily_limit_cents"`
	PerTransactionLimitCents  int64     `json:"per_transaction_limit_cents"`
	AllowedVendors            []string  `json:"allowed_vendors"`
	AllowedMCCs               []string  `json:"allowed_mccs"`
	AllowedWeekdaysUTC        []int     `json:"allowed_weekdays_utc"`
	AllowedHoursUTC           []int     `json:"allowed_hours_utc"`
	RequireApprovalAboveCents int64     `json:"require_approval_above_cents"`
	PurchaseGuideline         string    `json:"purchase_guideline"`
}
