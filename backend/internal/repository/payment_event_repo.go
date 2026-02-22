package repository

import (
	"context"
	"database/sql"
	"fmt"
)

// PaymentEventRepository stores processed webhook event IDs for idempotency.
type PaymentEventRepository struct {
	db *sql.DB
}

func NewPaymentEventRepository(db *sql.DB) *PaymentEventRepository {
	return &PaymentEventRepository{db: db}
}

// MarkProcessed inserts event ID if not seen. Returns true when inserted.
func (r *PaymentEventRepository) MarkProcessed(ctx context.Context, tx *sql.Tx, eventID string, provider string) (bool, error) {
	result, err := tx.ExecContext(ctx, `
		INSERT INTO payment_webhook_events (event_id, provider)
		VALUES ($1, $2)
		ON CONFLICT (event_id) DO NOTHING
	`, eventID, provider)
	if err != nil {
		return false, fmt.Errorf("failed to mark webhook event processed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to inspect webhook insert result: %w", err)
	}

	return rows > 0, nil
}
