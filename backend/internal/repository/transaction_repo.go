package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"agentpay/internal/models"

	"github.com/google/uuid"
)

// TransactionRepository handles transaction data persistence.
type TransactionRepository struct {
	db *sql.DB
}

var ErrTransactionNotFound = errors.New("transaction not found")

// NewTransactionRepository creates a new transaction repository.
func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// GetByRequestID retrieves a transaction by request_id (for idempotency).
func (r *TransactionRepository) GetByRequestID(ctx context.Context, requestID uuid.UUID) (*models.Transaction, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, request_id, agent_id, amount_cents, currency, vendor, status, reason,
		       provider, provider_session_id, provider_payment_intent_id, provider_status, provider_checkout_url,
		       meta, created_at
		FROM transactions
		WHERE request_id = $1
	`, requestID)

	txn, err := scanTransactionRow(row)
	if err != nil {
		if errors.Is(err, ErrTransactionNotFound) {
			return nil, nil // Not found is OK for idempotency check
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	return txn, nil
}

// Create creates a new transaction.
// Must be called within a transaction context.
func (r *TransactionRepository) Create(ctx context.Context, tx *sql.Tx, txn *models.Transaction) error {
	if txn.ID == uuid.Nil {
		txn.ID = uuid.New()
	}

	if txn.ProviderStatus == "" {
		txn.ProviderStatus = "not_applicable"
	}

	metaBytes, err := json.Marshal(txn.Meta)
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	err = tx.QueryRowContext(ctx, `
		INSERT INTO transactions (
			id, request_id, agent_id, amount_cents, currency, vendor, status, reason,
			provider, provider_session_id, provider_payment_intent_id, provider_status, provider_checkout_url,
			meta, created_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13,
			$14::jsonb, now()
		)
		ON CONFLICT (request_id) DO NOTHING
		RETURNING id, created_at
	`, txn.ID, txn.RequestID, txn.AgentID, txn.AmountCents, txn.Currency, txn.Vendor, txn.Status, txn.Reason,
		txn.Provider, nullableString(txn.ProviderSessionID), nullableString(txn.ProviderPaymentIntentID), txn.ProviderStatus, nullableString(txn.ProviderCheckoutURL),
		metaBytes).Scan(
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
	row := tx.QueryRowContext(ctx, `
		SELECT id, request_id, agent_id, amount_cents, currency, vendor, status, reason,
		       provider, provider_session_id, provider_payment_intent_id, provider_status, provider_checkout_url,
		       meta, created_at
		FROM transactions
		WHERE request_id = $1
	`, txn.RequestID)
	existingTxn, err := scanTransactionRow(row)
	if err != nil {
		return fmt.Errorf("failed to get existing transaction: %w", err)
	}
	*txn = *existingTxn

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

// HasApprovedVendorForAgent checks whether agent has prior approved spend with vendor.
func (r *TransactionRepository) HasApprovedVendorForAgent(ctx context.Context, tx *sql.Tx, agentID uuid.UUID, vendor string) (bool, error) {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM transactions
			WHERE agent_id = $1
			  AND status = 'APPROVED'
			  AND vendor = $2
		)
	`, agentID, strings.ToLower(strings.TrimSpace(vendor))).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check vendor history: %w", err)
	}
	return exists, nil
}

// GetApprovedAverageSpendForAgent returns average approved transaction amount in cents.
func (r *TransactionRepository) GetApprovedAverageSpendForAgent(ctx context.Context, tx *sql.Tx, agentID uuid.UUID) (int64, error) {
	var average sql.NullFloat64
	err := tx.QueryRowContext(ctx, `
		SELECT AVG(amount_cents)::float8
		FROM transactions
		WHERE agent_id = $1
		  AND status = 'APPROVED'
	`, agentID).Scan(&average)
	if err != nil {
		return 0, fmt.Errorf("failed to get average spend: %w", err)
	}
	if !average.Valid {
		return 0, nil
	}
	return int64(average.Float64), nil
}

// GetByID retrieves a transaction by ID.
func (r *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, request_id, agent_id, amount_cents, currency, vendor, status, reason,
		       provider, provider_session_id, provider_payment_intent_id, provider_status, provider_checkout_url,
		       meta, created_at
		FROM transactions
		WHERE id = $1
	`, id)
	return scanTransactionRow(row)
}

// GetByIDForUpdate retrieves and locks a transaction row by ID inside a transaction.
func (r *TransactionRepository) GetByIDForUpdate(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*models.Transaction, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT id, request_id, agent_id, amount_cents, currency, vendor, status, reason,
		       provider, provider_session_id, provider_payment_intent_id, provider_status, provider_checkout_url,
		       meta, created_at
		FROM transactions
		WHERE id = $1
		FOR UPDATE
	`, id)
	return scanTransactionRow(row)
}

// UpdateStatus updates a transaction status/reason and returns the updated record.
func (r *TransactionRepository) UpdateStatus(ctx context.Context, tx *sql.Tx, id uuid.UUID, status, reason string) (*models.Transaction, error) {
	row := tx.QueryRowContext(ctx, `
		UPDATE transactions
		SET status = $2, reason = $3
		WHERE id = $1
		RETURNING id, request_id, agent_id, amount_cents, currency, vendor, status, reason,
		          provider, provider_session_id, provider_payment_intent_id, provider_status, provider_checkout_url,
		          meta, created_at
	`, id, status, reason)
	return scanTransactionRow(row)
}

// UpdateStatusAndPayment updates a transaction decision and provider checkout fields in one statement.
func (r *TransactionRepository) UpdateStatusAndPayment(
	ctx context.Context,
	tx *sql.Tx,
	id uuid.UUID,
	status string,
	reason string,
	provider string,
	sessionID string,
	paymentIntentID string,
	providerStatus string,
	checkoutURL string,
) (*models.Transaction, error) {
	row := tx.QueryRowContext(ctx, `
		UPDATE transactions
		SET status = $2,
		    reason = $3,
		    provider = NULLIF($4, ''),
		    provider_session_id = NULLIF($5, ''),
		    provider_payment_intent_id = NULLIF($6, ''),
		    provider_status = $7,
		    provider_checkout_url = NULLIF($8, '')
		WHERE id = $1
		RETURNING id, request_id, agent_id, amount_cents, currency, vendor, status, reason,
		          provider, provider_session_id, provider_payment_intent_id, provider_status, provider_checkout_url,
		          meta, created_at
	`, id, status, reason, provider, sessionID, paymentIntentID, providerStatus, checkoutURL)
	return scanTransactionRow(row)
}

// ListByAgent retrieves recent transactions for an agent.
func (r *TransactionRepository) ListByAgent(ctx context.Context, agentID uuid.UUID, limit int) ([]models.Transaction, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, request_id, agent_id, amount_cents, currency, vendor, status, reason,
		       provider, provider_session_id, provider_payment_intent_id, provider_status, provider_checkout_url,
		       meta, created_at
		FROM transactions
		WHERE agent_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, agentID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions by agent: %w", err)
	}
	defer rows.Close()

	return scanTransactions(rows)
}

// List retrieves recent transactions, optionally filtered by agent.
func (r *TransactionRepository) List(ctx context.Context, agentID *uuid.UUID, limit int) ([]models.Transaction, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	var (
		rows *sql.Rows
		err  error
	)

	if agentID != nil {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, request_id, agent_id, amount_cents, currency, vendor, status, reason,
			       provider, provider_session_id, provider_payment_intent_id, provider_status, provider_checkout_url,
			       meta, created_at
			FROM transactions
			WHERE agent_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		`, *agentID, limit)
	} else {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, request_id, agent_id, amount_cents, currency, vendor, status, reason,
			       provider, provider_session_id, provider_payment_intent_id, provider_status, provider_checkout_url,
			       meta, created_at
			FROM transactions
			ORDER BY created_at DESC
			LIMIT $1
		`, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}
	defer rows.Close()

	return scanTransactions(rows)
}

// ListPending retrieves pending approval transactions.
func (r *TransactionRepository) ListPending(ctx context.Context, limit int) ([]models.Transaction, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, request_id, agent_id, amount_cents, currency, vendor, status, reason,
		       provider, provider_session_id, provider_payment_intent_id, provider_status, provider_checkout_url,
		       meta, created_at
		FROM transactions
		WHERE status = 'PENDING_APPROVAL'
		ORDER BY created_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending transactions: %w", err)
	}
	defer rows.Close()

	return scanTransactions(rows)
}

// UpdatePaymentState updates provider execution fields for a transaction.
func (r *TransactionRepository) UpdatePaymentState(
	ctx context.Context,
	txnID uuid.UUID,
	sessionID string,
	paymentIntentID string,
	providerStatus string,
) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE transactions
		SET provider_session_id = COALESCE(NULLIF($2, ''), provider_session_id),
		    provider_payment_intent_id = COALESCE(NULLIF($3, ''), provider_payment_intent_id),
		    provider_status = COALESCE(NULLIF($4, ''), provider_status)
		WHERE id = $1
	`, txnID, sessionID, paymentIntentID, providerStatus)
	if err != nil {
		return fmt.Errorf("failed to update payment state: %w", err)
	}
	return nil
}

func scanTransactions(rows *sql.Rows) ([]models.Transaction, error) {
	txs := make([]models.Transaction, 0)
	for rows.Next() {
		tx, err := scanTransactionRow(rows)
		if err != nil {
			return nil, err
		}
		txs = append(txs, *tx)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterating transactions: %w", err)
	}
	return txs, nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanTransactionRow(s scanner) (*models.Transaction, error) {
	txn := &models.Transaction{}
	var metaBytes []byte
	var provider sql.NullString
	var providerSessionID sql.NullString
	var providerPaymentIntentID sql.NullString
	var providerStatus sql.NullString
	var providerCheckoutURL sql.NullString
	err := s.Scan(
		&txn.ID,
		&txn.RequestID,
		&txn.AgentID,
		&txn.AmountCents,
		&txn.Currency,
		&txn.Vendor,
		&txn.Status,
		&txn.Reason,
		&provider,
		&providerSessionID,
		&providerPaymentIntentID,
		&providerStatus,
		&providerCheckoutURL,
		&metaBytes,
		&txn.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrTransactionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan transaction: %w", err)
	}
	if err := json.Unmarshal(metaBytes, &txn.Meta); err != nil {
		txn.Meta = make(map[string]interface{})
	}
	txn.Provider = provider.String
	txn.ProviderSessionID = providerSessionID.String
	txn.ProviderPaymentIntentID = providerPaymentIntentID.String
	txn.ProviderStatus = providerStatus.String
	txn.ProviderCheckoutURL = providerCheckoutURL.String
	return txn, nil
}

func nullableString(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
