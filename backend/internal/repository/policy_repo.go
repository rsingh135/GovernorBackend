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
		"allowed_vendors":              req.AllowedVendors,
		"require_approval_above_cents": req.RequireApprovalAboveCents,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal raw policy: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO policies (agent_id, daily_limit_cents, allowed_vendors, require_approval_above_cents, raw_policy, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5::jsonb, now(), now())
		ON CONFLICT (agent_id) DO UPDATE SET
			daily_limit_cents = EXCLUDED.daily_limit_cents,
			allowed_vendors = EXCLUDED.allowed_vendors,
			require_approval_above_cents = EXCLUDED.require_approval_above_cents,
			raw_policy = EXCLUDED.raw_policy,
			updated_at = now()
	`, req.AgentID, req.DailyLimitCents, pq.Array(req.AllowedVendors), req.RequireApprovalAboveCents, rawPolicy)

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

	err := r.db.QueryRowContext(ctx, `
		SELECT id, agent_id, daily_limit_cents, allowed_vendors, require_approval_above_cents, created_at, updated_at
		FROM policies
		WHERE agent_id = $1
	`, agentID).Scan(
		&policy.ID,
		&policy.AgentID,
		&policy.DailyLimitCents,
		pq.Array(&allowedVendors),
		&policy.RequireApprovalAboveCents,
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
	return policy, nil
}

// GetByAgentIDForUpdate retrieves a policy with row-level lock.
// Must be called within a transaction.
func (r *PolicyRepository) GetByAgentIDForUpdate(ctx context.Context, tx *sql.Tx, agentID uuid.UUID) (*models.Policy, error) {
	policy := &models.Policy{}
	var allowedVendors []string

	err := tx.QueryRowContext(ctx, `
		SELECT id, agent_id, daily_limit_cents, allowed_vendors, require_approval_above_cents, created_at, updated_at
		FROM policies
		WHERE agent_id = $1
		FOR UPDATE
	`, agentID).Scan(
		&policy.ID,
		&policy.AgentID,
		&policy.DailyLimitCents,
		pq.Array(&allowedVendors),
		&policy.RequireApprovalAboveCents,
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
	return policy, nil
}
