package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"agentpay/internal/models"
	"agentpay/internal/repository"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// validEvents is the set of supported webhook event types.
var validEvents = map[string]struct{}{
	"transaction.approved": {},
	"transaction.denied":   {},
}

// retryDelays defines the wait before each retry attempt (index = attempt number, 0-based).
var retryDelays = []time.Duration{
	0,               // attempt 1: immediate
	1 * time.Minute, // attempt 2
	5 * time.Minute, // attempt 3
	30 * time.Minute, // attempt 4
	2 * time.Hour,   // attempt 5
	24 * time.Hour,  // attempt 6 (final)
}

// WebhookService handles webhook registration and delivery.
type WebhookService struct {
	db          *sql.DB
	webhookRepo *repository.WebhookRepository
	httpClient  *http.Client
}

// NewWebhookService creates a new WebhookService.
func NewWebhookService(db *sql.DB) *WebhookService {
	return &WebhookService{
		db:          db,
		webhookRepo: repository.NewWebhookRepository(db),
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

// Register validates and creates a new webhook.
func (s *WebhookService) Register(ctx context.Context, agentID uuid.UUID, req *models.RegisterWebhookRequest) (*models.RegisterWebhookResponse, error) {
	if err := validateWebhookRequest(req); err != nil {
		return nil, err
	}

	w, err := s.webhookRepo.Create(ctx, agentID, req.URL, req.Secret, req.Events)
	if err != nil {
		return nil, err
	}

	return &models.RegisterWebhookResponse{
		ID:        w.ID,
		URL:       w.URL,
		Events:    w.Events,
		CreatedAt: w.CreatedAt,
	}, nil
}

// List returns all webhooks for an agent.
func (s *WebhookService) List(ctx context.Context, agentID uuid.UUID) (*models.ListWebhooksResponse, error) {
	webhooks, err := s.webhookRepo.ListByAgentID(ctx, agentID)
	if err != nil {
		return nil, err
	}
	return &models.ListWebhooksResponse{Webhooks: webhooks}, nil
}

// Delete removes a webhook owned by agentID.
func (s *WebhookService) Delete(ctx context.Context, webhookID uuid.UUID, agentID uuid.UUID) error {
	return s.webhookRepo.Delete(ctx, webhookID, agentID)
}

// EnqueueDeliveries creates delivery rows for all active webhooks subscribed to event.
// Must be called inside an existing DB transaction (tx).
func (s *WebhookService) EnqueueDeliveries(ctx context.Context, tx *sql.Tx, txn *models.Transaction, event string) error {
	webhooks, err := s.webhookRepo.FindActiveByAgentIDAndEvent(ctx, txn.AgentID, event)
	if err != nil {
		return err
	}
	if len(webhooks) == 0 {
		return nil
	}

	payload := models.WebhookEventPayload{
		EventID:       uuid.New(),
		Event:         event,
		TransactionID: txn.ID,
		AgentID:       txn.AgentID,
		AmountCents:   txn.AmountCents,
		Vendor:        txn.Vendor,
		Status:        strings.ToLower(txn.Status),
		Timestamp:     time.Now(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	for _, wh := range webhooks {
		if err := s.webhookRepo.CreateDelivery(ctx, tx, wh.ID, txn.ID, event, body); err != nil {
			return err
		}
	}

	return nil
}

// ProcessPendingDeliveries polls for due deliveries and fires them.
// Called by the background worker goroutine in main.go.
func (s *WebhookService) ProcessPendingDeliveries(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	deliveries, err := s.webhookRepo.PollPendingDeliveries(ctx, tx, 50)
	if err != nil {
		return err
	}

	for _, d := range deliveries {
		secret, err := s.webhookRepo.GetWebhookSecret(ctx, d.WebhookID)
		if err != nil {
			// Webhook deleted — cascade should have removed deliveries, but handle gracefully.
			log.Warn().Str("delivery_id", d.ID.String()).Msg("webhook not found for delivery, skipping")
			continue
		}

		httpStatus, deliverErr := s.deliver(ctx, d, secret)

		attempt := d.Attempt + 1
		var statusPtr *int
		if httpStatus > 0 {
			statusPtr = &httpStatus
		}

		if deliverErr == nil && httpStatus >= 200 && httpStatus < 300 {
			if err := s.webhookRepo.MarkDelivered(ctx, tx, d.ID, httpStatus); err != nil {
				log.Error().Err(err).Str("delivery_id", d.ID.String()).Msg("failed to mark delivered")
			}
			log.Info().
				Str("delivery_id", d.ID.String()).
				Str("event", d.Event).
				Int("http_status", httpStatus).
				Msg("webhook delivered")
		} else if attempt >= len(retryDelays) {
			// Exhausted all retries.
			if err := s.webhookRepo.MarkFailed(ctx, tx, d.ID, statusPtr); err != nil {
				log.Error().Err(err).Str("delivery_id", d.ID.String()).Msg("failed to mark failed")
			}
			log.Warn().
				Str("delivery_id", d.ID.String()).
				Str("event", d.Event).
				Msg("webhook delivery exhausted retries")
		} else {
			nextAt := time.Now().Add(retryDelays[attempt])
			if err := s.webhookRepo.ScheduleRetry(ctx, tx, d.ID, attempt, nextAt, statusPtr); err != nil {
				log.Error().Err(err).Str("delivery_id", d.ID.String()).Msg("failed to schedule retry")
			}
			log.Debug().
				Str("delivery_id", d.ID.String()).
				Int("attempt", attempt).
				Time("next_attempt_at", nextAt).
				Msg("webhook delivery scheduled for retry")
		}
	}

	return tx.Commit()
}

// deliver fires a single HTTP POST to the webhook URL, returning (httpStatus, error).
func (s *WebhookService) deliver(ctx context.Context, d *models.WebhookDelivery, secret string) (int, error) {
	sig := sign(secret, d.Payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "", bytes.NewReader(d.Payload))
	if err != nil {
		return 0, err
	}

	// Fetch URL from delivery's webhook (stored in payload context via webhook lookup above).
	// We need the URL — fetch it via the repo using a background context so it doesn't share the tx context.
	var webhookURL string
	err = s.db.QueryRowContext(ctx, `SELECT url FROM webhooks WHERE id = $1`, d.WebhookID).Scan(&webhookURL)
	if err != nil {
		return 0, fmt.Errorf("failed to get webhook url: %w", err)
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(d.Payload))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", sig)
	req.Header.Set("X-Webhook-Event", d.Event)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

// sign returns "sha256=<hex>" HMAC-SHA256 of body using secret.
func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// validateWebhookRequest checks URL, events, and secret.
func validateWebhookRequest(req *models.RegisterWebhookRequest) error {
	if req.URL == "" {
		return fmt.Errorf("url is required")
	}
	u, err := url.Parse(req.URL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("url must be a valid http or https URL")
	}

	if len(req.Events) == 0 {
		return fmt.Errorf("at least one event is required")
	}
	for _, e := range req.Events {
		if _, ok := validEvents[e]; !ok {
			return fmt.Errorf("unknown event: %s", e)
		}
	}

	if len(req.Secret) < 16 {
		return fmt.Errorf("secret must be at least 16 characters")
	}

	return nil
}
