package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"agentpay/internal/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// WebhookRepository handles webhook and delivery persistence.
type WebhookRepository struct {
	db *sql.DB
}

// NewWebhookRepository creates a new webhook repository.
func NewWebhookRepository(db *sql.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

// Create inserts a new webhook registration.
func (r *WebhookRepository) Create(ctx context.Context, agentID uuid.UUID, url, secret string, events []string) (*models.Webhook, error) {
	w := &models.Webhook{}
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO webhooks (agent_id, url, secret, events)
		VALUES ($1, $2, $3, $4)
		RETURNING id, agent_id, url, events, active, created_at
	`, agentID, url, secret, pq.Array(events)).Scan(
		&w.ID,
		&w.AgentID,
		&w.URL,
		pq.Array(&w.Events),
		&w.Active,
		&w.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}
	return w, nil
}

// ListByAgentID returns all webhooks for an agent (secret excluded).
func (r *WebhookRepository) ListByAgentID(ctx context.Context, agentID uuid.UUID) ([]models.Webhook, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, agent_id, url, events, active, created_at
		FROM webhooks
		WHERE agent_id = $1
		ORDER BY created_at DESC
	`, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}
	defer rows.Close()

	webhooks := []models.Webhook{}
	for rows.Next() {
		w := models.Webhook{}
		if err := rows.Scan(&w.ID, &w.AgentID, &w.URL, pq.Array(&w.Events), &w.Active, &w.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan webhook: %w", err)
		}
		webhooks = append(webhooks, w)
	}
	return webhooks, rows.Err()
}

// FindActiveByAgentIDAndEvent returns webhooks (with secret) matching agent + event.
func (r *WebhookRepository) FindActiveByAgentIDAndEvent(ctx context.Context, agentID uuid.UUID, event string) ([]models.WebhookWithSecret, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, agent_id, url, secret, events, active, created_at
		FROM webhooks
		WHERE agent_id = $1 AND active = TRUE AND $2 = ANY(events)
	`, agentID, event)
	if err != nil {
		return nil, fmt.Errorf("failed to find webhooks: %w", err)
	}
	defer rows.Close()

	result := []models.WebhookWithSecret{}
	for rows.Next() {
		var ws models.WebhookWithSecret
		if err := rows.Scan(
			&ws.ID, &ws.AgentID, &ws.URL, &ws.Secret,
			pq.Array(&ws.Events), &ws.Active, &ws.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan webhook: %w", err)
		}
		result = append(result, ws)
	}
	return result, rows.Err()
}

// Delete removes a webhook, enforcing ownership by agentID.
func (r *WebhookRepository) Delete(ctx context.Context, webhookID uuid.UUID, agentID uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM webhooks WHERE id = $1 AND agent_id = $2
	`, webhookID, agentID)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("webhook not found")
	}
	return nil
}

// CreateDelivery inserts a delivery row within an existing DB transaction.
func (r *WebhookRepository) CreateDelivery(ctx context.Context, tx *sql.Tx, webhookID, transactionID uuid.UUID, event string, payload []byte) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO webhook_deliveries (webhook_id, transaction_id, event, payload)
		VALUES ($1, $2, $3, $4::jsonb)
	`, webhookID, transactionID, event, payload)
	if err != nil {
		return fmt.Errorf("failed to create delivery: %w", err)
	}
	return nil
}

// PollPendingDeliveries fetches up to limit due deliveries with SKIP LOCKED.
// Must be called within a transaction.
func (r *WebhookRepository) PollPendingDeliveries(ctx context.Context, tx *sql.Tx, limit int) ([]*models.WebhookDelivery, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, webhook_id, transaction_id, event, payload, status, attempt, next_attempt_at, last_attempt_at, last_http_status, created_at
		FROM webhook_deliveries
		WHERE status = 'pending' AND next_attempt_at <= now()
		ORDER BY next_attempt_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to poll deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []*models.WebhookDelivery
	for rows.Next() {
		d := &models.WebhookDelivery{}
		if err := rows.Scan(
			&d.ID, &d.WebhookID, &d.TransactionID, &d.Event, &d.Payload,
			&d.Status, &d.Attempt, &d.NextAttemptAt, &d.LastAttemptAt,
			&d.LastHTTPStatus, &d.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan delivery: %w", err)
		}
		deliveries = append(deliveries, d)
	}
	return deliveries, rows.Err()
}

// MarkDelivered sets a delivery as successfully delivered.
func (r *WebhookRepository) MarkDelivered(ctx context.Context, tx *sql.Tx, id uuid.UUID, httpStatus int) error {
	now := time.Now()
	_, err := tx.ExecContext(ctx, `
		UPDATE webhook_deliveries
		SET status = 'delivered', last_attempt_at = $2, last_http_status = $3
		WHERE id = $1
	`, id, now, httpStatus)
	return err
}

// MarkFailed sets a delivery as permanently failed.
func (r *WebhookRepository) MarkFailed(ctx context.Context, tx *sql.Tx, id uuid.UUID, httpStatus *int) error {
	now := time.Now()
	_, err := tx.ExecContext(ctx, `
		UPDATE webhook_deliveries
		SET status = 'failed', last_attempt_at = $2, last_http_status = $3
		WHERE id = $1
	`, id, now, httpStatus)
	return err
}

// ScheduleRetry increments attempt count and sets the next retry time.
func (r *WebhookRepository) ScheduleRetry(ctx context.Context, tx *sql.Tx, id uuid.UUID, attempt int, nextAt time.Time, httpStatus *int) error {
	now := time.Now()
	_, err := tx.ExecContext(ctx, `
		UPDATE webhook_deliveries
		SET attempt = $2, next_attempt_at = $3, last_attempt_at = $4, last_http_status = $5
		WHERE id = $1
	`, id, attempt, nextAt, now, httpStatus)
	return err
}

// GetWebhookSecret returns just the secret for a given webhook ID (internal use).
func (r *WebhookRepository) GetWebhookSecret(ctx context.Context, webhookID uuid.UUID) (string, error) {
	var secret string
	err := r.db.QueryRowContext(ctx, `SELECT secret FROM webhooks WHERE id = $1`, webhookID).Scan(&secret)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("webhook not found")
	}
	return secret, err
}
