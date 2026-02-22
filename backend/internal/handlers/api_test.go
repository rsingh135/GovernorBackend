package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"agentpay/internal/middleware"
	"agentpay/internal/models"
	"agentpay/internal/services"
	"agentpay/internal/testutil"

	"github.com/google/uuid"
)

// TestSpendHandler_SuccessfulSpend tests a valid spend that should be approved.
func TestSpendHandler_SuccessfulSpend(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	ranveerUserID, _, bhangraBotAPIKey := testutil.SeedTestData(t, db)

	// Initialize services and handlers
	spendService := services.NewSpendService(db)
	agentService := services.NewAgentService(db)
	authMW := middleware.NewAuthMiddleware(agentService)
	handler := NewSpendHandler(spendService)

	// Create spend request (under daily limit, allowed vendor)
	reqBody := models.SpendRequest{
		RequestID: uuid.New(),
		Amount:    500, // $5.00 (under $20 daily limit)
		Vendor:    "openai.com",
		Meta:      map[string]interface{}{"test": "data"},
	}

	req := makeSpendRequest(t, reqBody, bhangraBotAPIKey)
	rr := httptest.NewRecorder()

	// Execute with auth middleware
	authMW.Authenticate(handler.Spend)(rr, req)

	// Assert response
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var resp models.SpendResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != "approved" {
		t.Errorf("Expected status 'approved', got '%s' with reason '%s'", resp.Status, resp.Reason)
	}

	// Verify balance was deducted
	finalBalance := testutil.GetUserBalance(t, db, ranveerUserID)
	expectedBalance := int64(5000 - 500) // Initial 5000 - spent 500
	if finalBalance != expectedBalance {
		t.Errorf("Expected final balance %d, got %d", expectedBalance, finalBalance)
	}

	t.Logf("✅ Successful spend test passed. Final balance: %d cents", finalBalance)
}

// TestSpendHandler_ExceedsDailyLimit tests spending that exceeds the daily limit.
func TestSpendHandler_ExceedsDailyLimit(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	ranveerUserID, _, bhangraBotAPIKey := testutil.SeedTestData(t, db)

	spendService := services.NewSpendService(db)
	agentService := services.NewAgentService(db)
	authMW := middleware.NewAuthMiddleware(agentService)
	handler := NewSpendHandler(spendService)

	// First spend: $15 (within daily limit of $20)
	req1 := models.SpendRequest{
		RequestID: uuid.New(),
		Amount:    1500,
		Vendor:    "openai.com",
		Meta:      map[string]interface{}{},
	}
	executeSpend(t, authMW, handler, req1, bhangraBotAPIKey, "approved")

	// Second spend: $10 (total would be $25, exceeds $20 limit)
	req2 := models.SpendRequest{
		RequestID: uuid.New(),
		Amount:    1000,
		Vendor:    "openai.com",
		Meta:      map[string]interface{}{},
	}

	req := makeSpendRequest(t, req2, bhangraBotAPIKey)
	rr := httptest.NewRecorder()
	authMW.Authenticate(handler.Spend)(rr, req)

	var resp models.SpendResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != "denied" {
		t.Errorf("Expected status 'denied', got '%s'", resp.Status)
	}

	if resp.Reason != "daily_limit_exceeded" {
		t.Errorf("Expected reason 'daily_limit_exceeded', got '%s'", resp.Reason)
	}

	// Verify balance only deducted for first transaction
	finalBalance := testutil.GetUserBalance(t, db, ranveerUserID)
	expectedBalance := int64(5000 - 1500) // Only first spend deducted
	if finalBalance != expectedBalance {
		t.Errorf("Expected final balance %d, got %d", expectedBalance, finalBalance)
	}

	t.Logf("✅ Daily limit exceeded test passed. Final balance: %d cents", finalBalance)
}

// TestSpendHandler_UnauthorizedVendor tests spending at a vendor not in allowed list.
func TestSpendHandler_UnauthorizedVendor(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	ranveerUserID, _, bhangraBotAPIKey := testutil.SeedTestData(t, db)

	spendService := services.NewSpendService(db)
	agentService := services.NewAgentService(db)
	authMW := middleware.NewAuthMiddleware(agentService)
	handler := NewSpendHandler(spendService)

	// Try to spend at unauthorized vendor
	reqBody := models.SpendRequest{
		RequestID: uuid.New(),
		Amount:    500,
		Vendor:    "stripe.com", // NOT in allowed_vendors
		Meta:      map[string]interface{}{},
	}

	req := makeSpendRequest(t, reqBody, bhangraBotAPIKey)
	rr := httptest.NewRecorder()
	authMW.Authenticate(handler.Spend)(rr, req)

	var resp models.SpendResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != "denied" {
		t.Errorf("Expected status 'denied', got '%s'", resp.Status)
	}

	if resp.Reason != "vendor_not_allowed" {
		t.Errorf("Expected reason 'vendor_not_allowed', got '%s'", resp.Reason)
	}

	// Verify balance NOT deducted
	finalBalance := testutil.GetUserBalance(t, db, ranveerUserID)
	if finalBalance != 5000 {
		t.Errorf("Expected balance unchanged at 5000, got %d", finalBalance)
	}

	t.Logf("✅ Unauthorized vendor test passed. Balance unchanged: %d cents", finalBalance)
}

// TestSpendHandler_Idempotency tests that duplicate request_id does not double-spend.
func TestSpendHandler_Idempotency(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	ranveerUserID, _, bhangraBotAPIKey := testutil.SeedTestData(t, db)

	spendService := services.NewSpendService(db)
	agentService := services.NewAgentService(db)
	authMW := middleware.NewAuthMiddleware(agentService)
	handler := NewSpendHandler(spendService)

	// Same request_id used twice
	requestID := uuid.New()
	reqBody := models.SpendRequest{
		RequestID: requestID,
		Amount:    800,
		Vendor:    "openai.com",
		Meta:      map[string]interface{}{"idempotency": "test"},
	}

	// First request
	req1 := makeSpendRequest(t, reqBody, bhangraBotAPIKey)
	rr1 := httptest.NewRecorder()
	authMW.Authenticate(handler.Spend)(rr1, req1)

	var resp1 models.SpendResponse
	if err := json.NewDecoder(rr1.Body).Decode(&resp1); err != nil {
		t.Fatalf("Failed to decode first response: %v", err)
	}

	if resp1.Status != "approved" {
		t.Errorf("First request: expected 'approved', got '%s'", resp1.Status)
	}

	balanceAfterFirst := testutil.GetUserBalance(t, db, ranveerUserID)

	// Second request with SAME request_id
	req2 := makeSpendRequest(t, reqBody, bhangraBotAPIKey)
	rr2 := httptest.NewRecorder()
	authMW.Authenticate(handler.Spend)(rr2, req2)

	var resp2 models.SpendResponse
	if err := json.NewDecoder(rr2.Body).Decode(&resp2); err != nil {
		t.Fatalf("Failed to decode second response: %v", err)
	}

	// Must return same result
	if resp2.Status != resp1.Status {
		t.Errorf("Idempotency violated: first status '%s', second status '%s'", resp1.Status, resp2.Status)
	}

	if resp2.Reason != resp1.Reason {
		t.Errorf("Idempotency violated: first reason '%s', second reason '%s'", resp1.Reason, resp2.Reason)
	}

	// Verify balance deducted only ONCE
	balanceAfterSecond := testutil.GetUserBalance(t, db, ranveerUserID)
	if balanceAfterFirst != balanceAfterSecond {
		t.Errorf("Balance changed after idempotent request! Before: %d, After: %d", balanceAfterFirst, balanceAfterSecond)
	}

	expectedBalance := int64(5000 - 800)
	if balanceAfterSecond != expectedBalance {
		t.Errorf("Expected final balance %d, got %d", expectedBalance, balanceAfterSecond)
	}

	// Verify only ONE transaction exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM transactions WHERE request_id = $1", requestID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count transactions: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected exactly 1 transaction, found %d", count)
	}

	t.Logf("✅ Idempotency test passed. Balance deducted only once: %d cents", balanceAfterSecond)
}

// TestSpendHandler_InsufficientBalance tests spending more than user balance.
func TestSpendHandler_InsufficientBalance(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	ranveerUserID, _, bhangraBotAPIKey := testutil.SeedTestData(t, db)

	spendService := services.NewSpendService(db)
	agentService := services.NewAgentService(db)
	authMW := middleware.NewAuthMiddleware(agentService)
	handler := NewSpendHandler(spendService)

	// Try to spend more than balance (balance is 5000 cents)
	reqBody := models.SpendRequest{
		RequestID: uuid.New(),
		Amount:    6000, // More than $50 balance
		Vendor:    "openai.com",
		Meta:      map[string]interface{}{},
	}

	req := makeSpendRequest(t, reqBody, bhangraBotAPIKey)
	rr := httptest.NewRecorder()
	authMW.Authenticate(handler.Spend)(rr, req)

	var resp models.SpendResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != "denied" {
		t.Errorf("Expected status 'denied', got '%s'", resp.Status)
	}

	if resp.Reason != "insufficient_balance" {
		t.Errorf("Expected reason 'insufficient_balance', got '%s'", resp.Reason)
	}

	// Verify balance unchanged
	finalBalance := testutil.GetUserBalance(t, db, ranveerUserID)
	if finalBalance != 5000 {
		t.Errorf("Expected balance unchanged at 5000, got %d", finalBalance)
	}

	t.Logf("✅ Insufficient balance test passed. Balance unchanged: %d cents", finalBalance)
}

func TestSpendHandler_PerTransactionLimitExceeded(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	ranveerUserID, _, bhangraBotAPIKey := testutil.SeedTestData(t, db)

	_, err := db.Exec(`UPDATE policies SET per_transaction_limit_cents = 700`)
	if err != nil {
		t.Fatalf("Failed to update per transaction limit: %v", err)
	}

	spendService := services.NewSpendService(db)
	agentService := services.NewAgentService(db)
	authMW := middleware.NewAuthMiddleware(agentService)
	handler := NewSpendHandler(spendService)

	reqBody := models.SpendRequest{
		RequestID: uuid.New(),
		Amount:    800,
		Vendor:    "openai.com",
		Meta:      map[string]interface{}{},
	}

	req := makeSpendRequest(t, reqBody, bhangraBotAPIKey)
	rr := httptest.NewRecorder()
	authMW.Authenticate(handler.Spend)(rr, req)

	var resp models.SpendResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != "denied" {
		t.Errorf("Expected status 'denied', got '%s'", resp.Status)
	}
	if resp.Reason != "per_transaction_limit_exceeded" {
		t.Errorf("Expected reason 'per_transaction_limit_exceeded', got '%s'", resp.Reason)
	}

	finalBalance := testutil.GetUserBalance(t, db, ranveerUserID)
	if finalBalance != 5000 {
		t.Errorf("Expected balance unchanged at 5000, got %d", finalBalance)
	}
}

func TestSpendHandler_MCCNotAllowed(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	ranveerUserID, _, bhangraBotAPIKey := testutil.SeedTestData(t, db)

	_, err := db.Exec(`UPDATE policies SET allowed_mccs = ARRAY['5734']::text[]`)
	if err != nil {
		t.Fatalf("Failed to update MCC allowlist: %v", err)
	}

	spendService := services.NewSpendService(db)
	agentService := services.NewAgentService(db)
	authMW := middleware.NewAuthMiddleware(agentService)
	handler := NewSpendHandler(spendService)

	reqBody := models.SpendRequest{
		RequestID: uuid.New(),
		Amount:    500,
		Vendor:    "openai.com",
		MCC:       "5815",
		Meta:      map[string]interface{}{},
	}

	req := makeSpendRequest(t, reqBody, bhangraBotAPIKey)
	rr := httptest.NewRecorder()
	authMW.Authenticate(handler.Spend)(rr, req)

	var resp models.SpendResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != "denied" {
		t.Errorf("Expected status 'denied', got '%s'", resp.Status)
	}
	if resp.Reason != "mcc_not_allowed" {
		t.Errorf("Expected reason 'mcc_not_allowed', got '%s'", resp.Reason)
	}

	finalBalance := testutil.GetUserBalance(t, db, ranveerUserID)
	if finalBalance != 5000 {
		t.Errorf("Expected balance unchanged at 5000, got %d", finalBalance)
	}
}

func TestSpendHandler_OutsideAllowedTimeWindow(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	ranveerUserID, _, bhangraBotAPIKey := testutil.SeedTestData(t, db)

	now := time.Now().UTC()
	nonMatchingDay := (int(now.Weekday()) + 1) % 7
	_, err := db.Exec(`UPDATE policies SET allowed_weekdays_utc = ARRAY[$1]::integer[]`, nonMatchingDay)
	if err != nil {
		t.Fatalf("Failed to update allowed weekdays: %v", err)
	}

	spendService := services.NewSpendService(db)
	agentService := services.NewAgentService(db)
	authMW := middleware.NewAuthMiddleware(agentService)
	handler := NewSpendHandler(spendService)

	reqBody := models.SpendRequest{
		RequestID: uuid.New(),
		Amount:    500,
		Vendor:    "openai.com",
		Meta:      map[string]interface{}{},
	}

	req := makeSpendRequest(t, reqBody, bhangraBotAPIKey)
	rr := httptest.NewRecorder()
	authMW.Authenticate(handler.Spend)(rr, req)

	var resp models.SpendResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != "denied" {
		t.Errorf("Expected status 'denied', got '%s'", resp.Status)
	}
	if resp.Reason != "outside_allowed_time_window" {
		t.Errorf("Expected reason 'outside_allowed_time_window', got '%s'", resp.Reason)
	}

	finalBalance := testutil.GetUserBalance(t, db, ranveerUserID)
	if finalBalance != 5000 {
		t.Errorf("Expected balance unchanged at 5000, got %d", finalBalance)
	}
}

func TestSpendHandler_OrganizationFrozen(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	ranveerUserID, _, bhangraBotAPIKey := testutil.SeedTestData(t, db)

	_, err := db.Exec(`UPDATE users SET status = 'frozen'`)
	if err != nil {
		t.Fatalf("Failed to freeze user: %v", err)
	}

	spendService := services.NewSpendService(db)
	agentService := services.NewAgentService(db)
	authMW := middleware.NewAuthMiddleware(agentService)
	handler := NewSpendHandler(spendService)

	reqBody := models.SpendRequest{
		RequestID: uuid.New(),
		Amount:    500,
		Vendor:    "openai.com",
		Meta:      map[string]interface{}{},
	}

	req := makeSpendRequest(t, reqBody, bhangraBotAPIKey)
	rr := httptest.NewRecorder()
	authMW.Authenticate(handler.Spend)(rr, req)

	var resp models.SpendResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != "denied" {
		t.Errorf("Expected status 'denied', got '%s'", resp.Status)
	}
	if resp.Reason != "organization_frozen" {
		t.Errorf("Expected reason 'organization_frozen', got '%s'", resp.Reason)
	}

	finalBalance := testutil.GetUserBalance(t, db, ranveerUserID)
	if finalBalance != 5000 {
		t.Errorf("Expected balance unchanged at 5000, got %d", finalBalance)
	}
}

// TestSpendHandler_GuidelineMismatch tests denying spend when vendor doesn't match policy guideline.
func TestSpendHandler_GuidelineMismatch(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	ranveerUserID, _, bhangraBotAPIKey := testutil.SeedTestData(t, db)

	_, err := db.Exec(`UPDATE policies SET purchase_guideline = 'legal filings and accounting services only'`)
	if err != nil {
		t.Fatalf("Failed to update purchase guideline: %v", err)
	}

	spendService := services.NewSpendService(db)
	agentService := services.NewAgentService(db)
	authMW := middleware.NewAuthMiddleware(agentService)
	handler := NewSpendHandler(spendService)

	reqBody := models.SpendRequest{
		RequestID: uuid.New(),
		Amount:    500,
		Vendor:    "openai.com",
		Meta:      map[string]interface{}{},
	}

	req := makeSpendRequest(t, reqBody, bhangraBotAPIKey)
	rr := httptest.NewRecorder()
	authMW.Authenticate(handler.Spend)(rr, req)

	var resp models.SpendResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != "denied" {
		t.Errorf("Expected status 'denied', got '%s'", resp.Status)
	}
	if resp.Reason != "guideline_mismatch" {
		t.Errorf("Expected reason 'guideline_mismatch', got '%s'", resp.Reason)
	}

	finalBalance := testutil.GetUserBalance(t, db, ranveerUserID)
	if finalBalance != 5000 {
		t.Errorf("Expected balance unchanged at 5000, got %d", finalBalance)
	}
}

// TestSpendHandler_GuidelineRelevant tests allowing spend when guideline aligns with vendor domain intent.
func TestSpendHandler_GuidelineRelevant(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	ranveerUserID, _, bhangraBotAPIKey := testutil.SeedTestData(t, db)

	_, err := db.Exec(`UPDATE policies SET purchase_guideline = 'AI and LLM tooling for engineering productivity'`)
	if err != nil {
		t.Fatalf("Failed to update purchase guideline: %v", err)
	}

	spendService := services.NewSpendService(db)
	agentService := services.NewAgentService(db)
	authMW := middleware.NewAuthMiddleware(agentService)
	handler := NewSpendHandler(spendService)

	reqBody := models.SpendRequest{
		RequestID: uuid.New(),
		Amount:    500,
		Vendor:    "openai.com",
		Meta:      map[string]interface{}{},
	}

	req := makeSpendRequest(t, reqBody, bhangraBotAPIKey)
	rr := httptest.NewRecorder()
	authMW.Authenticate(handler.Spend)(rr, req)

	var resp models.SpendResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != "approved" {
		t.Errorf("Expected status 'approved', got '%s' with reason '%s'", resp.Status, resp.Reason)
	}

	finalBalance := testutil.GetUserBalance(t, db, ranveerUserID)
	if finalBalance != 4500 {
		t.Errorf("Expected balance 4500 after approved spend, got %d", finalBalance)
	}
}

// Helper functions

func makeSpendRequest(t *testing.T, reqBody models.SpendRequest, apiKey string) *http.Request {
	t.Helper()

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/spend", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apiKey", apiKey)

	return req
}

func executeSpend(t *testing.T, authMW *middleware.AuthMiddleware, handler *SpendHandler, reqBody models.SpendRequest, apiKey string, expectedStatus string) {
	t.Helper()

	req := makeSpendRequest(t, reqBody, apiKey)
	rr := httptest.NewRecorder()
	authMW.Authenticate(handler.Spend)(rr, req)

	var resp models.SpendResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != expectedStatus {
		t.Fatalf("Expected status '%s', got '%s' with reason '%s'", expectedStatus, resp.Status, resp.Reason)
	}
}
