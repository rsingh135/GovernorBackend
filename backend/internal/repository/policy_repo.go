package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"agentpay/internal/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// PolicyRepository handles policy data persistence.
type PolicyRepository struct {
	db *sql.DB
}

// NewPolicyRepository creates a new policy repository.
func NewPolicyRepository(db *sql.DB) *PolicyRepository {
	return &PolicyRepository{db: db}
}

// Upsert creates or updates a policy for an agent.
func (r *PolicyRepository) Upsert(ctx context.Context, req *models.UpsertPolicyRequest) (*models.Policy, error) {
	rawPolicy, err := json.Marshal(map[string]interface{}{
		"daily_limit_cents":            req.DailyLimitCents,
		"per_transaction_limit_cents":  req.PerTransactionLimitCents,
		"allowed_vendors":              req.AllowedVendors,
		"allowed_mccs":                 req.AllowedMCCs,
		"allowed_weekdays_utc":         req.AllowedWeekdaysUTC,
		"allowed_hours_utc":            req.AllowedHoursUTC,
		"require_approval_above_cents": req.RequireApprovalAboveCents,
		"purchase_guideline":           req.PurchaseGuideline,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal raw policy: %w", err)
	}

	allowedWeekdays := toInt64Slice(req.AllowedWeekdaysUTC)
	allowedHours := toInt64Slice(req.AllowedHoursUTC)

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO policies (
			agent_id, daily_limit_cents, per_transaction_limit_cents,
			allowed_vendors, allowed_mccs, allowed_weekdays_utc, allowed_hours_utc,
			require_approval_above_cents, purchase_guideline, raw_policy, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb, now(), now())
		ON CONFLICT (agent_id) DO UPDATE SET
			daily_limit_cents = EXCLUDED.daily_limit_cents,
			per_transaction_limit_cents = EXCLUDED.per_transaction_limit_cents,
			allowed_vendors = EXCLUDED.allowed_vendors,
			allowed_mccs = EXCLUDED.allowed_mccs,
			allowed_weekdays_utc = EXCLUDED.allowed_weekdays_utc,
			allowed_hours_utc = EXCLUDED.allowed_hours_utc,
			require_approval_above_cents = EXCLUDED.require_approval_above_cents,
			purchase_guideline = EXCLUDED.purchase_guideline,
			raw_policy = EXCLUDED.raw_policy,
			updated_at = now()
	`, req.AgentID, req.DailyLimitCents, req.PerTransactionLimitCents, pq.Array(req.AllowedVendors), pq.Array(req.AllowedMCCs), pq.Array(allowedWeekdays), pq.Array(allowedHours), req.RequireApprovalAboveCents, req.PurchaseGuideline, rawPolicy)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			return nil, fmt.Errorf("agent not found")
		}
		return nil, fmt.Errorf("failed to upsert policy: %w", err)
	}

	// Fetch the created/updated policy
	return r.GetByAgentID(ctx, req.AgentID)
}

// GetByAgentID retrieves a policy by agent ID.
func (r *PolicyRepository) GetByAgentID(ctx context.Context, agentID uuid.UUID) (*models.Policy, error) {
	policy := &models.Policy{}
	var allowedVendors []string
	var allowedMCCs []string
	var allowedWeekdays []int64
	var allowedHours []int64

	err := r.db.QueryRowContext(ctx, `
		SELECT id, agent_id, daily_limit_cents, per_transaction_limit_cents,
		       allowed_vendors, allowed_mccs, allowed_weekdays_utc, allowed_hours_utc,
		       require_approval_above_cents, purchase_guideline, created_at, updated_at
		FROM policies
		WHERE agent_id = $1
	`, agentID).Scan(
		&policy.ID,
		&policy.AgentID,
		&policy.DailyLimitCents,
		&policy.PerTransactionLimitCents,
		pq.Array(&allowedVendors),
		pq.Array(&allowedMCCs),
		pq.Array(&allowedWeekdays),
		pq.Array(&allowedHours),
		&policy.RequireApprovalAboveCents,
		&policy.PurchaseGuideline,
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("policy not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	policy.AllowedVendors = allowedVendors
	policy.AllowedMCCs = allowedMCCs
	policy.AllowedWeekdaysUTC = toIntSlice(allowedWeekdays)
	policy.AllowedHoursUTC = toIntSlice(allowedHours)
	return policy, nil
}

// GetByAgentIDForUpdate retrieves a policy with row-level lock.
// Must be called within a transaction.
func (r *PolicyRepository) GetByAgentIDForUpdate(ctx context.Context, tx *sql.Tx, agentID uuid.UUID) (*models.Policy, error) {
	policy := &models.Policy{}
	var allowedVendors []string
	var allowedMCCs []string
	var allowedWeekdays []int64
	var allowedHours []int64

	err := tx.QueryRowContext(ctx, `
		SELECT id, agent_id, daily_limit_cents, per_transaction_limit_cents,
		       allowed_vendors, allowed_mccs, allowed_weekdays_utc, allowed_hours_utc,
		       require_approval_above_cents, purchase_guideline, created_at, updated_at
		FROM policies
		WHERE agent_id = $1
		FOR UPDATE
	`, agentID).Scan(
		&policy.ID,
		&policy.AgentID,
		&policy.DailyLimitCents,
		&policy.PerTransactionLimitCents,
		pq.Array(&allowedVendors),
		pq.Array(&allowedMCCs),
		pq.Array(&allowedWeekdays),
		pq.Array(&allowedHours),
		&policy.RequireApprovalAboveCents,
		&policy.PurchaseGuideline,
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("policy not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to lock policy: %w", err)
	}

	policy.AllowedVendors = allowedVendors
	policy.AllowedMCCs = allowedMCCs
	policy.AllowedWeekdaysUTC = toIntSlice(allowedWeekdays)
	policy.AllowedHoursUTC = toIntSlice(allowedHours)
	return policy, nil
}

func toInt64Slice(input []int) []int64 {
	out := make([]int64, 0, len(input))
	for _, v := range input {
		out = append(out, int64(v))
	}
	return out
}

func toIntSlice(input []int64) []int {
	out := make([]int, 0, len(input))
	for _, v := range input {
		out = append(out, int(v))
	}
	return out
}
