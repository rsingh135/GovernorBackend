package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"agentpay/internal/models"
	"agentpay/internal/repository"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// ApprovalService handles approve/deny workflow for pending transactions.
type ApprovalService struct {
	db             *sql.DB
	userRepo       *repository.UserRepository
	agentRepo      *repository.AgentRepository
	txnRepo        *repository.TransactionRepository
	approvalRepo   *repository.ApprovalRepository
	webhookService *WebhookService
}

// NewApprovalService creates a new approval service.
func NewApprovalService(db *sql.DB, webhookService *WebhookService) *ApprovalService {
	return &ApprovalService{
		db:             db,
		userRepo:       repository.NewUserRepository(db),
		agentRepo:      repository.NewAgentRepository(db),
		txnRepo:        repository.NewTransactionRepository(db),
		approvalRepo:   repository.NewApprovalRepository(db),
		webhookService: webhookService,
	}
}

// Approve approves a PENDING_APPROVAL transaction and deducts the balance.
func (s *ApprovalService) Approve(ctx context.Context, txnID uuid.UUID, approverUserID uuid.UUID) (*models.ApproveResponse, error) {
	return s.processDecision(ctx, txnID, approverUserID, "APPROVED")
}

// Deny denies a PENDING_APPROVAL transaction.
func (s *ApprovalService) Deny(ctx context.Context, txnID uuid.UUID, approverUserID uuid.UUID) (*models.ApproveResponse, error) {
	return s.processDecision(ctx, txnID, approverUserID, "DENIED")
}

// processDecision handles the shared approve/deny logic.
func (s *ApprovalService) processDecision(ctx context.Context, txnID uuid.UUID, approverUserID uuid.UUID, newStatus string) (*models.ApproveResponse, error) {
	// Verify approver exists and has permission (outside tx for performance)
	approver, err := s.userRepo.GetByID(ctx, approverUserID)
	if err != nil {
		return nil, fmt.Errorf("approver not found")
	}
	if !approver.CanApprove {
		return nil, fmt.Errorf("user does not have approval permission")
	}

	// Verify transaction exists
	existing, err := s.txnRepo.GetByID(ctx, txnID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("transaction not found")
	}
	if existing.Status != "PENDING_APPROVAL" {
		return nil, fmt.Errorf("transaction is not pending approval (status: %s)", existing.Status)
	}

	// Begin DB transaction with 5s timeout
	txCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(txCtx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Lock transaction row and re-check status
	txn, err := s.txnRepo.GetByIDForUpdate(txCtx, tx, txnID)
	if err != nil {
		return nil, err
	}
	if txn == nil {
		return nil, fmt.Errorf("transaction not found")
	}
	if txn.Status != "PENDING_APPROVAL" {
		return nil, fmt.Errorf("transaction is not pending approval (status: %s)", txn.Status)
	}

	// For approval: lock user row and re-verify balance
	if newStatus == "APPROVED" {
		agent, err := s.agentRepo.GetByID(txCtx, txn.AgentID)
		if err != nil {
			return nil, fmt.Errorf("agent not found")
		}

		user, err := s.userRepo.GetByIDForUpdate(txCtx, tx, agent.UserID)
		if err != nil {
			return nil, err
		}
		if user.BalanceCents < txn.AmountCents {
			return nil, fmt.Errorf("insufficient balance")
		}

		if err := s.userRepo.DeductBalance(txCtx, tx, agent.UserID, txn.AmountCents); err != nil {
			return nil, err
		}
	}

	// Update transaction status
	if err := s.txnRepo.UpdateApproval(txCtx, tx, txnID, newStatus, approverUserID); err != nil {
		return nil, err
	}

	// Record audit action
	action := "approved"
	if newStatus == "DENIED" {
		action = "denied"
	}
	if _, err := s.approvalRepo.Create(txCtx, tx, txnID, approverUserID, action); err != nil {
		return nil, err
	}

	// Enqueue webhook deliveries atomically inside the same tx
	event := "transaction." + action
	if err := s.webhookService.EnqueueDeliveries(txCtx, tx, txn, event); err != nil {
		return nil, fmt.Errorf("failed to enqueue webhook deliveries: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	now := time.Now()
	resp := &models.ApproveResponse{
		TransactionID:    txnID,
		Status:           newStatus,
		ApprovedByUserID: &approverUserID,
	}
	if newStatus == "APPROVED" {
		resp.ApprovedAt = &now
	}

	log.Info().
		Str("transaction_id", txnID.String()).
		Str("approver_user_id", approverUserID.String()).
		Str("action", action).
		Int64("amount_cents", txn.AmountCents).
		Msg("transaction decision recorded")

	return resp, nil
}
