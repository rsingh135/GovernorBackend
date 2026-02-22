package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"agentpay/internal/middleware"
	"agentpay/internal/models"
	"agentpay/internal/payments"
	"agentpay/internal/repository"
	"agentpay/internal/services"
	"agentpay/internal/testutil"

	"github.com/google/uuid"
)

type fakePaymentProvider struct {
	enabled bool
}

func (f *fakePaymentProvider) Enabled() bool { return f.enabled }
func (f *fakePaymentProvider) Name() string  { return "fakepay" }

func (f *fakePaymentProvider) CreateCheckoutSession(_ context.Context, req payments.CreateCheckoutRequest) (*payments.CheckoutSession, error) {
	if !f.enabled {
		return nil, fmt.Errorf("provider disabled")
	}
	return &payments.CheckoutSession{
		Provider:        f.Name(),
		SessionID:       "cs_test_" + req.TransactionID.String()[:8],
		CheckoutURL:     "https://checkout.test/session/" + req.TransactionID.String(),
		PaymentIntentID: "pi_test_" + req.TransactionID.String()[:8],
		ProviderStatus:  "checkout_created",
	}, nil
}

func (f *fakePaymentProvider) ParseWebhook(_ []byte, _ string) (*payments.WebhookEvent, error) {
	return nil, fmt.Errorf("not implemented")
}

func TestSpendHandler_AutoApprovedCreatesCheckoutSession(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	_, _, bhangraBotAPIKey := testutil.SeedTestData(t, db)

	provider := &fakePaymentProvider{enabled: true}
	spendService := services.NewSpendServiceWithProvider(db, provider)
	agentService := services.NewAgentService(db)
	authMW := middleware.NewAuthMiddleware(agentService)
	handler := NewSpendHandler(spendService)

	reqBody := models.SpendRequest{
		RequestID: uuid.New(),
		Amount:    500,
		Vendor:    "openai.com",
		Meta:      map[string]interface{}{"phase": "2"},
	}

	req := makeSpendRequest(t, reqBody, bhangraBotAPIKey)
	rr := httptest.NewRecorder()
	authMW.Authenticate(handler.Spend)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var resp models.SpendResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed decoding spend response: %v", err)
	}

	if resp.Status != "approved" {
		t.Fatalf("expected approved status, got %s", resp.Status)
	}
	if resp.CheckoutURL == "" {
		t.Fatalf("expected checkout_url to be populated")
	}
	if resp.ProviderStatus != "checkout_created" {
		t.Fatalf("expected provider_status checkout_created, got %s", resp.ProviderStatus)
	}

	txnRepo := repository.NewTransactionRepository(db)
	txn, err := txnRepo.GetByRequestID(context.Background(), reqBody.RequestID)
	if err != nil {
		t.Fatalf("failed to fetch transaction by request id: %v", err)
	}
	if txn == nil {
		t.Fatalf("expected transaction to exist")
	}
	if txn.Provider != "fakepay" {
		t.Fatalf("expected provider fakepay, got %s", txn.Provider)
	}
	if txn.ProviderCheckoutURL == "" {
		t.Fatalf("expected provider checkout url on transaction")
	}
}

func TestAdminDashboard_ApprovePendingCreatesCheckoutSession(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	ranveerUserID, _, bhangraBotAPIKey := testutil.SeedTestData(t, db)
	pendingTxnID := createPendingTransaction(t, db, bhangraBotAPIKey)

	provider := &fakePaymentProvider{enabled: true}
	adminAuthService := services.NewAdminAuthService(db, 24*time.Hour)
	adminMW := middleware.NewAdminAuthMiddleware(adminAuthService)
	dashboardService := services.NewAdminDashboardServiceWithProvider(db, provider)
	dashboardHandler := NewAdminDashboardHandler(dashboardService)
	token := adminLoginToken(t, db)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/admin/transactions/%s/approve", pendingTxnID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	adminMW.Authenticate(dashboardHandler.ApproveTransaction)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var body map[string]models.Transaction
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode approve response: %v", err)
	}
	txn := body["transaction"]

	if txn.Status != "approved" {
		t.Fatalf("expected approved status, got %s", txn.Status)
	}
	if txn.ProviderCheckoutURL == "" {
		t.Fatalf("expected provider checkout url after approval")
	}
	if txn.ProviderStatus != "checkout_created" {
		t.Fatalf("expected provider_status checkout_created, got %s", txn.ProviderStatus)
	}

	balance := testutil.GetUserBalance(t, db, ranveerUserID)
	if balance != 3400 {
		t.Fatalf("expected balance 3400 after pending approval, got %d", balance)
	}
}

func TestPaymentWebhookService_IdempotentUpdate(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	_, agentID, _ := testutil.SeedTestData(t, db)
	requestID := uuid.New()
	testutil.CreateTestTransaction(t, db, &models.Transaction{
		RequestID:   requestID,
		AgentID:     agentID,
		AmountCents: 500,
		Vendor:      "openai.com",
		Status:      "APPROVED",
		Reason:      "approved",
	})

	txnRepo := repository.NewTransactionRepository(db)
	txn, err := txnRepo.GetByRequestID(context.Background(), requestID)
	if err != nil || txn == nil {
		t.Fatalf("failed to fetch transaction: %v", err)
	}

	webhookService := services.NewPaymentWebhookService(db)
	event := &payments.WebhookEvent{
		EventID:         "evt_test_once",
		Provider:        "stripe",
		TransactionID:   txn.ID,
		ProviderStatus:  "payment_succeeded",
		SessionID:       "cs_webhook_123",
		PaymentIntentID: "pi_webhook_123",
	}

	if err := webhookService.HandleEvent(context.Background(), event); err != nil {
		t.Fatalf("first webhook handle failed: %v", err)
	}
	if err := webhookService.HandleEvent(context.Background(), event); err != nil {
		t.Fatalf("second webhook handle failed: %v", err)
	}

	updated, err := txnRepo.GetByID(context.Background(), txn.ID)
	if err != nil {
		t.Fatalf("failed to load updated transaction: %v", err)
	}
	if updated.ProviderStatus != "payment_succeeded" {
		t.Fatalf("expected provider_status payment_succeeded, got %s", updated.ProviderStatus)
	}
	if updated.ProviderSessionID != "cs_webhook_123" {
		t.Fatalf("expected session id to be set")
	}
	if updated.ProviderPaymentIntentID != "pi_webhook_123" {
		t.Fatalf("expected payment intent id to be set")
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM payment_webhook_events WHERE event_id = $1", "evt_test_once").Scan(&count); err != nil {
		t.Fatalf("failed counting webhook events: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 webhook event row, got %d", count)
	}
}
