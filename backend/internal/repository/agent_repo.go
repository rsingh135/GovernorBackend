package repository

import (
	"context"
	"database/sql"
	"fmt"

	"agentpay/internal/apikey"
	"agentpay/internal/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// AgentRepository handles agent data persistence.
type AgentRepository struct {
	db *sql.DB
}

// NewAgentRepository creates a new agent repository.
func NewAgentRepository(db *sql.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

// Create creates a new agent with API key.
func (r *AgentRepository) Create(ctx context.Context, userID uuid.UUID, name string) (*models.Agent, string, error) {
	agentID := uuid.New()

	// Generate secure API key
	var apiKeyStr string
	var err error

	for i := 0; i < 3; i++ {
		var keyHash []byte
		var prefix string
		apiKeyStr, keyHash, prefix, err = apikey.Generate()
		if err != nil {
			return nil, "", fmt.Errorf("failed to generate api key: %w", err)
		}

		_, err = r.db.ExecContext(ctx, `
			INSERT INTO agents (id, user_id, name, status, api_key_hash, api_key_prefix, created_at)
			VALUES ($1, $2, $3, 'active', $4, $5, now())
		`, agentID, userID, name, keyHash, prefix)

		if err == nil {
			break
		}

		// Retry on unique constraint violation (extremely unlikely)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			continue
		}

		return nil, "", fmt.Errorf("failed to create agent: %w", err)
	}

	if err != nil {
		return nil, "", fmt.Errorf("failed to create agent after retries: %w", err)
	}

	agent := &models.Agent{
		ID:           agentID,
		UserID:       userID,
		Name:         name,
		Status:       "active",
		APIKeyPrefix: apiKeyStr[:16],
	}

	return agent, apiKeyStr, nil
}

// GetByAPIKey retrieves an agent by API key hash.
func (r *AgentRepository) GetByAPIKey(ctx context.Context, apiKeyStr string) (*models.Agent, error) {
	keyHash := apikey.Hash(apiKeyStr)

	agent := &models.Agent{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, status, api_key_prefix, created_at
		FROM agents
		WHERE api_key_hash = $1
	`, keyHash).Scan(
		&agent.ID,
		&agent.UserID,
		&agent.Name,
		&agent.Status,
		&agent.APIKeyPrefix,
		&agent.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid api key")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	return agent, nil
}

// GetByID retrieves an agent by ID.
func (r *AgentRepository) GetByID(ctx context.Context, agentID uuid.UUID) (*models.Agent, error) {
	agent := &models.Agent{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, status, api_key_prefix, created_at
		FROM agents
		WHERE id = $1
	`, agentID).Scan(
		&agent.ID,
		&agent.UserID,
		&agent.Name,
		&agent.Status,
		&agent.APIKeyPrefix,
		&agent.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("agent not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	return agent, nil
}

// LockAgentForUpdate locks an agent row for update within a transaction.
func (r *AgentRepository) LockAgentForUpdate(ctx context.Context, tx *sql.Tx, agentID uuid.UUID) (string, error) {
	var status string
	err := tx.QueryRowContext(ctx, `
		SELECT status
		FROM agents
		WHERE id = $1
		FOR UPDATE
	`, agentID).Scan(&status)

	if err == sql.ErrNoRows {
		return "", fmt.Errorf("agent not found")
	}
	if err != nil {
		return "", fmt.Errorf("failed to lock agent: %w", err)
	}

	return status, nil
}

// List retrieves paginated agents with filters.
func (r *AgentRepository) List(
	ctx context.Context,
	filters models.AgentFilters,
	pagination models.PaginationParams,
) ([]models.Agent, error) {
	query := `
		SELECT id, user_id, name, status, api_key_prefix, created_at
		FROM agents
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	// Apply filters
	if filters.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, *filters.UserID)
		argIdx++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *filters.Status)
		argIdx++
	}

	// Order by newest first
	query += " ORDER BY created_at DESC"

	// Apply pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, pagination.Limit, pagination.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	defer rows.Close()

	agents := []models.Agent{}
	for rows.Next() {
		agent := models.Agent{}
		err := rows.Scan(
			&agent.ID,
			&agent.UserID,
			&agent.Name,
			&agent.Status,
			&agent.APIKeyPrefix,
			&agent.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}
		agents = append(agents, agent)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return agents, nil
}

// Count returns total count of agents matching filters.
func (r *AgentRepository) Count(
	ctx context.Context,
	filters models.AgentFilters,
) (int64, error) {
	query := `SELECT COUNT(*) FROM agents WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	// Apply same filters as List
	if filters.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, *filters.UserID)
		argIdx++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *filters.Status)
		argIdx++
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count agents: %w", err)
	}

	return count, nil
}
