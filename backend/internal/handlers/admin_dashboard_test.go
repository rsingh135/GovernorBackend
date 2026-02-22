package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"agentpay/internal/middleware"
	"agentpay/internal/models"
	"agentpay/internal/repository"
	"agentpay/internal/services"
	"agentpay/internal/testutil"

	"github.com/google/uuid"
)

func TestAdminDashboard_ApprovePendingTransaction(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	ranveerUserID, bhangraBotAgentID, bhangraBotAPIKey := testutil.SeedTestData(t, db)
	pendingTxnID := createPendingTransaction(t, db, bhangraBotAPIKey)

	adminMW, dashboardHandler := setupAdminDashboardHandlers(t, db)
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

	txn, ok := body["transaction"]
	if !ok {
		t.Fatalf("missing transaction in response")
	}
	if txn.Status != "approved" {
		t.Fatalf("expected transaction status approved, got %s", txn.Status)
	}
	if txn.AgentID != bhangraBotAgentID {
		t.Fatalf("expected agent id %s, got %s", bhangraBotAgentID, txn.AgentID)
	}

	balance := testutil.GetUserBalance(t, db, ranveerUserID)
	if balance != 3400 {
		t.Fatalf("expected user balance 3400 after approval, got %d", balance)
	}
}

func TestAdminDashboard_DenyPendingTransaction(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	ranveerUserID, _, bhangraBotAPIKey := testutil.SeedTestData(t, db)
	pendingTxnID := createPendingTransaction(t, db, bhangraBotAPIKey)

	adminMW, dashboardHandler := setupAdminDashboardHandlers(t, db)
	token := adminLoginToken(t, db)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/admin/transactions/%s/deny", pendingTxnID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	adminMW.Authenticate(dashboardHandler.DenyTransaction)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var body map[string]models.Transaction
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode deny response: %v", err)
	}

	txn, ok := body["transaction"]
	if !ok {
		t.Fatalf("missing transaction in response")
	}
	if txn.Status != "denied" {
		t.Fatalf("expected transaction status denied, got %s", txn.Status)
	}
	if txn.Reason != "human_denied" {
		t.Fatalf("expected denial reason human_denied, got %s", txn.Reason)
	}

	balance := testutil.GetUserBalance(t, db, ranveerUserID)
	if balance != 5000 {
		t.Fatalf("expected user balance unchanged at 5000, got %d", balance)
	}
}

func TestAdminDashboard_PendingQueueAndHistory(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	_, bhangraBotAgentID, bhangraBotAPIKey := testutil.SeedTestData(t, db)
	_ = createPendingTransaction(t, db, bhangraBotAPIKey)

	adminMW, dashboardHandler := setupAdminDashboardHandlers(t, db)
	token := adminLoginToken(t, db)

	pendingReq := httptest.NewRequest(http.MethodGet, "/admin/transactions/pending?limit=10", nil)
	pendingReq.Header.Set("Authorization", "Bearer "+token)
	pendingRR := httptest.NewRecorder()
	adminMW.Authenticate(dashboardHandler.ListPendingTransactions)(pendingRR, pendingReq)

	if pendingRR.Code != http.StatusOK {
		t.Fatalf("expected pending status 200, got %d body=%s", pendingRR.Code, pendingRR.Body.String())
	}

	var pendingBody map[string][]models.Transaction
	if err := json.NewDecoder(pendingRR.Body).Decode(&pendingBody); err != nil {
		t.Fatalf("failed to decode pending response: %v", err)
	}
	if len(pendingBody["transactions"]) == 0 {
		t.Fatalf("expected at least one pending transaction")
	}

	historyReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/agents/%s/history?limit=10", bhangraBotAgentID), nil)
	historyReq.Header.Set("Authorization", "Bearer "+token)
	historyRR := httptest.NewRecorder()
	adminMW.Authenticate(dashboardHandler.GetAgentHistory)(historyRR, historyReq)

	if historyRR.Code != http.StatusOK {
		t.Fatalf("expected history status 200, got %d body=%s", historyRR.Code, historyRR.Body.String())
	}

	var historyBody map[string][]models.Transaction
	if err := json.NewDecoder(historyRR.Body).Decode(&historyBody); err != nil {
		t.Fatalf("failed to decode history response: %v", err)
	}

	if len(historyBody["transactions"]) == 0 {
		t.Fatalf("expected at least one history transaction")
	}
}

func setupAdminDashboardHandlers(t *testing.T, db *sql.DB) (*middleware.AdminAuthMiddleware, *AdminDashboardHandler) {
	t.Helper()

	adminAuthService := services.NewAdminAuthService(db, 24*time.Hour)
	adminMW := middleware.NewAdminAuthMiddleware(adminAuthService)
	dashboardService := services.NewAdminDashboardService(db)
	dashboardHandler := NewAdminDashboardHandler(dashboardService)
	return adminMW, dashboardHandler
}

func adminLoginToken(t *testing.T, db *sql.DB) string {
	t.Helper()
	adminAuthService := services.NewAdminAuthService(db, 24*time.Hour)
	resp, err := adminAuthService.Login(context.Background(), "admin@governor.local", "governor_admin_123")
	if err != nil {
		t.Fatalf("failed to login admin: %v", err)
	}
	return resp.Token
}

func createPendingTransaction(t *testing.T, db *sql.DB, apiKey string) uuid.UUID {
	t.Helper()

	spendService := services.NewSpendService(db)
	agentService := services.NewAgentService(db)
	authMW := middleware.NewAuthMiddleware(agentService)
	handler := NewSpendHandler(spendService)

	reqBody := models.SpendRequest{
		RequestID: uuid.New(),
		Amount:    1600,
		Vendor:    "openai.com",
		Meta:      map[string]interface{}{"purpose": "approval-test"},
	}
	req := makeSpendRequest(t, reqBody, apiKey)
	rr := httptest.NewRecorder()
	authMW.Authenticate(handler.Spend)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected spend 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var spendResp models.SpendResponse
	if err := json.NewDecoder(rr.Body).Decode(&spendResp); err != nil {
		t.Fatalf("failed to decode spend response: %v", err)
	}
	if spendResp.Status != "pending_approval" {
		t.Fatalf("expected pending_approval, got %s", spendResp.Status)
	}

	txnRepo := repository.NewTransactionRepository(db)
	txn, err := txnRepo.GetByRequestID(context.Background(), reqBody.RequestID)
	if err != nil {
		t.Fatalf("failed to load transaction: %v", err)
	}
	if txn == nil {
		t.Fatalf("expected pending transaction to exist")
	}
	return txn.ID
}
