package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"agentpay/internal/models"

	"github.com/google/uuid"
)

// TransactionRepository handles transaction data persistence.
type TransactionRepository struct {
	db *sql.DB
}

// NewTransactionRepository creates a new transaction repository.
func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// GetByRequestID retrieves a transaction by request_id (for idempotency).
func (r *TransactionRepository) GetByRequestID(ctx context.Context, requestID uuid.UUID) (*models.Transaction, error) {
	txn := &models.Transaction{}
	var metaBytes []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, request_id, agent_id, amount_cents, currency, vendor, status, reason, meta, created_at
		FROM transactions
		WHERE request_id = $1
	`, requestID).Scan(
		&txn.ID,
		&txn.RequestID,
		&txn.AgentID,
		&txn.AmountCents,
		&txn.Currency,
		&txn.Vendor,
		&txn.Status,
		&txn.Reason,
		&metaBytes,
		&txn.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found is OK for idempotency check
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if err := json.Unmarshal(metaBytes, &txn.Meta); err != nil {
		txn.Meta = make(map[string]interface{})
	}

	return txn, nil
}

// Create creates a new transaction.
// Must be called within a transaction context.
func (r *TransactionRepository) Create(ctx context.Context, tx *sql.Tx, txn *models.Transaction) error {
	metaBytes, err := json.Marshal(txn.Meta)
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	err = tx.QueryRowContext(ctx, `
		INSERT INTO transactions (request_id, agent_id, amount_cents, currency, vendor, status, reason, meta, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, now())
		ON CONFLICT (request_id) DO NOTHING
		RETURNING id, created_at
	`, txn.RequestID, txn.AgentID, txn.AmountCents, txn.Currency, txn.Vendor, txn.Status, txn.Reason, metaBytes).Scan(
		&txn.ID,
		&txn.CreatedAt,
	)

	if err == sql.ErrNoRows {
		// Conflict: fetch existing transaction
		return r.getByRequestIDInTx(ctx, tx, txn)
	}
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// getByRequestIDInTx retrieves a transaction by request_id within a transaction.
func (r *TransactionRepository) getByRequestIDInTx(ctx context.Context, tx *sql.Tx, txn *models.Transaction) error {
	var metaBytes []byte

	err := tx.QueryRowContext(ctx, `
		SELECT id, request_id, agent_id, amount_cents, currency, vendor, status, reason, meta, created_at
		FROM transactions
		WHERE request_id = $1
	`, txn.RequestID).Scan(
		&txn.ID,
		&txn.RequestID,
		&txn.AgentID,
		&txn.AmountCents,
		&txn.Currency,
		&txn.Vendor,
		&txn.Status,
		&txn.Reason,
		&metaBytes,
		&txn.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to get existing transaction: %w", err)
	}

	if err := json.Unmarshal(metaBytes, &txn.Meta); err != nil {
		txn.Meta = make(map[string]interface{})
	}

	return nil
}

// GetTodaySpendForAgent calculates today's total approved spend for an agent.
// Must be called within a transaction.
func (r *TransactionRepository) GetTodaySpendForAgent(ctx context.Context, tx *sql.Tx, agentID uuid.UUID) (int64, error) {
	var totalSpent int64

	err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount_cents), 0)
		FROM transactions
		WHERE agent_id = $1
		  AND status = 'APPROVED'
		  AND created_at >= date_trunc('day', now())
		  AND created_at < date_trunc('day', now()) + interval '1 day'
	`, agentID).Scan(&totalSpent)

	if err != nil {
		return 0, fmt.Errorf("failed to calculate today's spend: %w", err)
	}

	return totalSpent, nil
}
