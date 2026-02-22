package services

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"agentpay/internal/models"
	"agentpay/internal/payments"
	"agentpay/internal/repository"

	"github.com/google/uuid"
)

// SpendService handles the core spending engine logic.
type SpendService struct {
	db         *sql.DB
	userRepo   *repository.UserRepository
	agentRepo  *repository.AgentRepository
	policyRepo *repository.PolicyRepository
	txnRepo    *repository.TransactionRepository
	payments   payments.Provider
}

// NewSpendService creates a new spend service.
func NewSpendService(db *sql.DB) *SpendService {
	return NewSpendServiceWithProvider(db, payments.NewNoopProvider())
}

// NewSpendServiceWithProvider creates a new spend service with a payment provider.
func NewSpendServiceWithProvider(db *sql.DB, provider payments.Provider) *SpendService {
	if provider == nil {
		provider = payments.NewNoopProvider()
	}

	return &SpendService{
		db:         db,
		userRepo:   repository.NewUserRepository(db),
		agentRepo:  repository.NewAgentRepository(db),
		policyRepo: repository.NewPolicyRepository(db),
		txnRepo:    repository.NewTransactionRepository(db),
		payments:   provider,
	}
}

// ProcessSpend is the core spending engine.
// CRITICAL: Uses SELECT FOR UPDATE for concurrency safety.
func (s *SpendService) ProcessSpend(ctx context.Context, agent *models.Agent, req *models.SpendRequest) (*models.SpendResponse, error) {
	// Normalize inputs
	req.Vendor = strings.ToLower(strings.TrimSpace(req.Vendor))
	req.MCC = strings.ToLower(strings.TrimSpace(req.MCC))
	if req.Meta == nil {
		req.Meta = make(map[string]interface{})
	}
	if req.MCC != "" {
		req.Meta["mcc"] = req.MCC
	}

	// Idempotency check (outside transaction for performance)
	existingTxn, err := s.txnRepo.GetByRequestID(ctx, req.RequestID)
	if err != nil {
		return nil, err
	}
	if existingTxn != nil {
		return s.transactionToResponse(existingTxn), nil
	}

	// Begin transaction with 5-second timeout
	txCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(txCtx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Re-check idempotency inside transaction (handle race conditions)
	existingTxn, err = s.txnRepo.GetByRequestID(txCtx, req.RequestID)
	if err != nil {
		return nil, err
	}
	if existingTxn != nil {
		_ = tx.Commit()
		return s.transactionToResponse(existingTxn), nil
	}

	// Lock agent row (serializes per-agent spending)
	agentStatus, err := s.agentRepo.LockAgentForUpdate(txCtx, tx, agent.ID)
	if err != nil {
		return nil, err
	}
	if agentStatus == "frozen" {
		txn := s.createDeniedTransaction(agent.ID, req, "agent_frozen")
		if err := s.txnRepo.Create(txCtx, tx, txn); err != nil {
			return nil, err
		}
		_ = tx.Commit()
		return s.transactionToResponse(txn), nil
	}

	// Lock user row and verify balance
	user, err := s.userRepo.GetByIDForUpdate(txCtx, tx, agent.UserID)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(user.Status, "frozen") {
		txn := s.createDeniedTransaction(agent.ID, req, "organization_frozen")
		if err := s.txnRepo.Create(txCtx, tx, txn); err != nil {
			return nil, err
		}
		_ = tx.Commit()
		return s.transactionToResponse(txn), nil
	}
	if user.BalanceCents < req.Amount {
		txn := s.createDeniedTransaction(agent.ID, req, "insufficient_balance")
		if err := s.txnRepo.Create(txCtx, tx, txn); err != nil {
			return nil, err
		}
		_ = tx.Commit()
		return s.transactionToResponse(txn), nil
	}

	// Lock policy row
	policy, err := s.policyRepo.GetByAgentIDForUpdate(txCtx, tx, agent.ID)
	if err != nil {
		txn := s.createDeniedTransaction(agent.ID, req, "no_policy")
		if err := s.txnRepo.Create(txCtx, tx, txn); err != nil {
			return nil, err
		}
		_ = tx.Commit()
		return s.transactionToResponse(txn), nil
	}

	// Calculate today's spend (serialized by agent lock)
	todaySpent, err := s.txnRepo.GetTodaySpendForAgent(txCtx, tx, agent.ID)
	if err != nil {
		return nil, err
	}

	// Evaluate spending rules
	if req.Amount > policy.PerTransactionLimitCents {
		txn := s.createDeniedTransaction(agent.ID, req, "per_transaction_limit_exceeded")
		if err := s.txnRepo.Create(txCtx, tx, txn); err != nil {
			return nil, err
		}
		_ = tx.Commit()
		return s.transactionToResponse(txn), nil
	}

	if req.Amount+todaySpent > policy.DailyLimitCents {
		txn := s.createDeniedTransaction(agent.ID, req, "daily_limit_exceeded")
		if err := s.txnRepo.Create(txCtx, tx, txn); err != nil {
			return nil, err
		}
		_ = tx.Commit()
		return s.transactionToResponse(txn), nil
	}

	if !s.isVendorAllowed(req.Vendor, policy.AllowedVendors) {
		txn := s.createDeniedTransaction(agent.ID, req, "vendor_not_allowed")
		if err := s.txnRepo.Create(txCtx, tx, txn); err != nil {
			return nil, err
		}
		_ = tx.Commit()
		return s.transactionToResponse(txn), nil
	}

	if !s.isMCCAllowed(req.MCC, policy.AllowedMCCs) {
		txn := s.createDeniedTransaction(agent.ID, req, "mcc_not_allowed")
		if err := s.txnRepo.Create(txCtx, tx, txn); err != nil {
			return nil, err
		}
		_ = tx.Commit()
		return s.transactionToResponse(txn), nil
	}

	if !s.isWithinAllowedWindow(time.Now().UTC(), policy.AllowedWeekdaysUTC, policy.AllowedHoursUTC) {
		txn := s.createDeniedTransaction(agent.ID, req, "outside_allowed_time_window")
		if err := s.txnRepo.Create(txCtx, tx, txn); err != nil {
			return nil, err
		}
		_ = tx.Commit()
		return s.transactionToResponse(txn), nil
	}

	if !s.isVendorRelevantToGuideline(req.Vendor, policy.PurchaseGuideline) {
		txn := s.createDeniedTransaction(agent.ID, req, "guideline_mismatch")
		if err := s.txnRepo.Create(txCtx, tx, txn); err != nil {
			return nil, err
		}
		_ = tx.Commit()
		return s.transactionToResponse(txn), nil
	}

	// Check approval threshold
	var status, reason string
	if policy.RequireApprovalAboveCents > 0 && req.Amount > policy.RequireApprovalAboveCents {
		status = "PENDING_APPROVAL"
		reason = "requires_approval"
	} else {
		status = "APPROVED"
		reason = "approved"
	}

	if status == "APPROVED" {
		averageSpend, err := s.txnRepo.GetApprovedAverageSpendForAgent(txCtx, tx, agent.ID)
		if err != nil {
			return nil, err
		}
		seenVendor, err := s.txnRepo.HasApprovedVendorForAgent(txCtx, tx, agent.ID, req.Vendor)
		if err != nil {
			return nil, err
		}
		if averageSpend > 0 && !seenVendor {
			status = "PENDING_APPROVAL"
			reason = "new_vendor_requires_approval"
		}
		if s.isUnusualSpend(req.Amount, averageSpend) {
			status = "PENDING_APPROVAL"
			reason = "unusual_spend_requires_approval"
		}
	}

	// Create transaction
	txn := &models.Transaction{
		ID:             uuid.New(),
		RequestID:      req.RequestID,
		AgentID:        agent.ID,
		AmountCents:    req.Amount,
		Currency:       "usd", // Default currency
		Vendor:         req.Vendor,
		Status:         status,
		Reason:         reason,
		Meta:           req.Meta,
		ProviderStatus: "not_applicable",
	}

	if status == "PENDING_APPROVAL" {
		txn.ProviderStatus = "awaiting_human_approval"
	}

	// For auto-approved flows, create a checkout session before persisting approval.
	// This keeps balance and transaction state consistent if checkout creation fails.
	if status == "APPROVED" && s.payments.Enabled() {
		checkout, err := s.payments.CreateCheckoutSession(txCtx, payments.CreateCheckoutRequest{
			TransactionID: txn.ID,
			AgentID:       agent.ID,
			AmountCents:   req.Amount,
			Currency:      txn.Currency,
			Vendor:        req.Vendor,
		})
		if err != nil {
			deniedTxn := s.createDeniedTransaction(agent.ID, req, "checkout_creation_failed")
			deniedTxn.ID = txn.ID
			deniedTxn.Provider = s.payments.Name()
			deniedTxn.ProviderStatus = "checkout_creation_failed"
			if err := s.txnRepo.Create(txCtx, tx, deniedTxn); err != nil {
				return nil, err
			}
			_ = tx.Commit()
			return s.transactionToResponse(deniedTxn), nil
		}

		txn.Provider = checkout.Provider
		txn.ProviderSessionID = checkout.SessionID
		txn.ProviderPaymentIntentID = checkout.PaymentIntentID
		txn.ProviderCheckoutURL = checkout.CheckoutURL
		txn.ProviderStatus = checkout.ProviderStatus
	}

	if err := s.txnRepo.Create(txCtx, tx, txn); err != nil {
		return nil, err
	}

	// Deduct from user balance (only for approved transactions)
	if status == "APPROVED" {
		if err := s.userRepo.DeductBalance(txCtx, tx, agent.UserID, req.Amount); err != nil {
			return nil, err
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.transactionToResponse(txn), nil
}

// createDeniedTransaction creates a denied transaction.
func (s *SpendService) createDeniedTransaction(agentID uuid.UUID, req *models.SpendRequest, reason string) *models.Transaction {
	return &models.Transaction{
		RequestID:   req.RequestID,
		AgentID:     agentID,
		AmountCents: req.Amount,
		Currency:    "usd", // Default currency
		Vendor:      req.Vendor,
		Status:      "DENIED",
		Reason:      reason,
		Meta:        req.Meta,
	}
}

// transactionToResponse converts a transaction to a spend response.
func (s *SpendService) transactionToResponse(txn *models.Transaction) *models.SpendResponse {
	// Convert DB status (uppercase) to API status (lowercase)
	status := strings.ToLower(strings.ReplaceAll(txn.Status, "_", "_"))
	return &models.SpendResponse{
		Status:         status,
		Reason:         txn.Reason,
		CheckoutURL:    txn.ProviderCheckoutURL,
		ProviderStatus: txn.ProviderStatus,
		TransactionID:  txn.ID.String(),
	}
}

// isVendorAllowed checks if a vendor is in the allowed list.
func (s *SpendService) isVendorAllowed(vendor string, allowedVendors []string) bool {
	if vendor == "" {
		return false
	}
	for _, allowed := range allowedVendors {
		if strings.ToLower(strings.TrimSpace(allowed)) == vendor {
			return true
		}
	}
	return false
}

// isMCCAllowed enforces optional MCC allowlists.
func (s *SpendService) isMCCAllowed(mcc string, allowedMCCs []string) bool {
	if len(allowedMCCs) == 0 {
		return true
	}
	mcc = strings.ToLower(strings.TrimSpace(mcc))
	if mcc == "" {
		return false
	}
	for _, allowed := range allowedMCCs {
		if strings.ToLower(strings.TrimSpace(allowed)) == mcc {
			return true
		}
	}
	return false
}

// isWithinAllowedWindow enforces optional UTC weekday/hour allowlists.
func (s *SpendService) isWithinAllowedWindow(now time.Time, allowedWeekdays []int, allowedHours []int) bool {
	weekdayAllowed := true
	if len(allowedWeekdays) > 0 {
		weekdayAllowed = false
		currentDay := int(now.UTC().Weekday())
		for _, day := range allowedWeekdays {
			if day == currentDay {
				weekdayAllowed = true
				break
			}
		}
	}

	if !weekdayAllowed {
		return false
	}

	hourAllowed := true
	if len(allowedHours) > 0 {
		hourAllowed = false
		currentHour := now.UTC().Hour()
		for _, hour := range allowedHours {
			if hour == currentHour {
				hourAllowed = true
				break
			}
		}
	}

	return hourAllowed
}

func (s *SpendService) isUnusualSpend(amountCents int64, averageCents int64) bool {
	if averageCents <= 0 {
		return false
	}
	// Trigger human review when spend is materially above baseline.
	return amountCents >= (averageCents*3) && amountCents >= 1000
}

var tokenSplitRegex = regexp.MustCompile(`[^a-z0-9]+`)

func (s *SpendService) isVendorRelevantToGuideline(vendor string, guideline string) bool {
	vendor = strings.ToLower(strings.TrimSpace(vendor))
	guideline = strings.ToLower(strings.TrimSpace(guideline))

	if guideline == "" {
		return true
	}
	if vendor == "" {
		return false
	}

	vendorCore := strings.ReplaceAll(vendor, ".", "")
	if strings.Contains(guideline, vendorCore) {
		return true
	}

	guidelineTokens := tokenSet(guideline)
	if len(guidelineTokens) == 0 {
		return true
	}

	for token := range tokenSet(vendor) {
		if len(token) < 2 {
			continue
		}
		if _, ok := guidelineTokens[token]; ok {
			return true
		}
		if len(token) >= 4 && strings.Contains(guideline, token) {
			return true
		}
	}

	for _, hint := range vendorGuidelineHints(vendor) {
		if _, ok := guidelineTokens[hint]; ok {
			return true
		}
	}

	return false
}

func tokenSet(input string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, raw := range tokenSplitRegex.Split(strings.ToLower(input), -1) {
		token := strings.TrimSpace(raw)
		if token == "" {
			continue
		}
		out[token] = struct{}{}
	}
	return out
}

func vendorGuidelineHints(vendor string) []string {
	hints := make([]string, 0, 4)

	addHints := func(words ...string) {
		hints = append(hints, words...)
	}

	switch {
	case strings.Contains(vendor, "openai"), strings.Contains(vendor, "anthropic"), strings.Contains(vendor, "claude"), strings.Contains(vendor, "perplexity"):
		addHints("ai", "llm", "model", "inference")
	case strings.Contains(vendor, "github"), strings.Contains(vendor, "gitlab"), strings.Contains(vendor, "bitbucket"):
		addHints("code", "developer", "engineering", "repository")
	case strings.Contains(vendor, "aws"), strings.Contains(vendor, "azure"), strings.Contains(vendor, "gcp"), strings.Contains(vendor, "vercel"), strings.Contains(vendor, "render"), strings.Contains(vendor, "railway"), strings.Contains(vendor, "digitalocean"):
		addHints("cloud", "infrastructure", "hosting", "compute")
	case strings.Contains(vendor, "slack"), strings.Contains(vendor, "notion"), strings.Contains(vendor, "atlassian"):
		addHints("productivity", "operations", "collaboration")
	case strings.Contains(vendor, "stripe"), strings.Contains(vendor, "paypal"):
		addHints("payments", "billing", "finance")
	}

	return hints
}
