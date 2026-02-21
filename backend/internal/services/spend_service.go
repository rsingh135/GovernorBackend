package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"agentpay/internal/models"
	"agentpay/internal/repository"

	"github.com/google/uuid"
)

// SpendService handles the core spending engine logic.
type SpendService struct {
	db          *sql.DB
	userRepo    *repository.UserRepository
	agentRepo   *repository.AgentRepository
	policyRepo  *repository.PolicyRepository
	txnRepo     *repository.TransactionRepository
}

// NewSpendService creates a new spend service.
func NewSpendService(db *sql.DB) *SpendService {
	return &SpendService{
		db:         db,
		userRepo:   repository.NewUserRepository(db),
		agentRepo:  repository.NewAgentRepository(db),
		policyRepo: repository.NewPolicyRepository(db),
		txnRepo:    repository.NewTransactionRepository(db),
	}
}

// ProcessSpend is the core spending engine.
// CRITICAL: Uses SELECT FOR UPDATE for concurrency safety.
func (s *SpendService) ProcessSpend(ctx context.Context, agent *models.Agent, req *models.SpendRequest) (*models.SpendResponse, error) {
	// Normalize inputs
	req.Vendor = strings.ToLower(strings.TrimSpace(req.Vendor))
	if req.Meta == nil {
		req.Meta = make(map[string]interface{})
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

	// Check approval threshold
	var status, reason string
	if policy.RequireApprovalAboveCents > 0 && req.Amount > policy.RequireApprovalAboveCents {
		status = "PENDING_APPROVAL"
		reason = "requires_approval"
	} else {
		status = "APPROVED"
		reason = "approved"
	}

	// Create transaction
	txn := &models.Transaction{
		RequestID:   req.RequestID,
		AgentID:     agent.ID,
		AmountCents: req.Amount,
		Currency:    "usd", // Default currency
		Vendor:      req.Vendor,
		Status:      status,
		Reason:      reason,
		Meta:        req.Meta,
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
		Status: status,
		Reason: txn.Reason,
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
