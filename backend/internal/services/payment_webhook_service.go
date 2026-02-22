package services

import (
	"context"
	"database/sql"
	"fmt"

	"agentpay/internal/payments"
	"agentpay/internal/repository"
)

// PaymentWebhookService processes normalized webhook events idempotently.
type PaymentWebhookService struct {
	db        *sql.DB
	txnRepo   *repository.TransactionRepository
	eventRepo *repository.PaymentEventRepository
}

func NewPaymentWebhookService(db *sql.DB) *PaymentWebhookService {
	return &PaymentWebhookService{
		db:        db,
		txnRepo:   repository.NewTransactionRepository(db),
		eventRepo: repository.NewPaymentEventRepository(db),
	}
}

// HandleEvent marks event idempotently and updates transaction payment state.
func (s *PaymentWebhookService) HandleEvent(ctx context.Context, event *payments.WebhookEvent) error {
	if event == nil {
		return fmt.Errorf("webhook event is nil")
	}
	if event.EventID == "" {
		return fmt.Errorf("webhook event id is required")
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("failed to begin webhook transaction: %w", err)
	}
	defer tx.Rollback()

	inserted, err := s.eventRepo.MarkProcessed(ctx, tx, event.EventID, event.Provider)
	if err != nil {
		return err
	}
	if !inserted {
		_ = tx.Commit()
		return nil
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE transactions
		SET provider_session_id = COALESCE(NULLIF($2, ''), provider_session_id),
		    provider_payment_intent_id = COALESCE(NULLIF($3, ''), provider_payment_intent_id),
		    provider_status = COALESCE(NULLIF($4, ''), provider_status)
		WHERE id = $1
	`, event.TransactionID, event.SessionID, event.PaymentIntentID, event.ProviderStatus)
	if err != nil {
		return fmt.Errorf("failed to update transaction from webhook: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit webhook transaction: %w", err)
	}

	return nil
}
