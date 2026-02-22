package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"agentpay/internal/models"
	"agentpay/internal/payments"
	"agentpay/internal/repository"

	"github.com/google/uuid"
)

// AdminDashboardService serves admin-facing read and review flows.
type AdminDashboardService struct {
	db         *sql.DB
	userRepo   *repository.UserRepository
	agentRepo  *repository.AgentRepository
	policyRepo *repository.PolicyRepository
	txnRepo    *repository.TransactionRepository
	auditRepo  *repository.AuditRepository
	payments   payments.Provider
}

// NewAdminDashboardService creates a new dashboard service.
func NewAdminDashboardService(db *sql.DB) *AdminDashboardService {
	return NewAdminDashboardServiceWithProvider(db, payments.NewNoopProvider())
}

// NewAdminDashboardServiceWithProvider creates a dashboard service with payment integration.
func NewAdminDashboardServiceWithProvider(db *sql.DB, provider payments.Provider) *AdminDashboardService {
	if provider == nil {
		provider = payments.NewNoopProvider()
	}

	return &AdminDashboardService{
		db:         db,
		userRepo:   repository.NewUserRepository(db),
		agentRepo:  repository.NewAgentRepository(db),
		policyRepo: repository.NewPolicyRepository(db),
		txnRepo:    repository.NewTransactionRepository(db),
		auditRepo:  repository.NewAuditRepository(db),
		payments:   provider,
	}
}

func (s *AdminDashboardService) ListUsers(ctx context.Context, limit int) ([]models.User, error) {
	return s.userRepo.List(ctx, limit)
}

func (s *AdminDashboardService) GetUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *AdminDashboardService) ListAgents(ctx context.Context, userID *uuid.UUID, limit int) ([]models.Agent, error) {
	return s.agentRepo.List(ctx, userID, limit)
}

func (s *AdminDashboardService) GetAgent(ctx context.Context, agentID uuid.UUID) (*models.Agent, error) {
	return s.agentRepo.GetByID(ctx, agentID)
}

func (s *AdminDashboardService) GetPolicyByAgent(ctx context.Context, agentID uuid.UUID) (*models.Policy, error) {
	return s.policyRepo.GetByAgentID(ctx, agentID)
}

func (s *AdminDashboardService) ListTransactions(ctx context.Context, agentID *uuid.UUID, limit int) ([]models.Transaction, error) {
	return s.txnRepo.List(ctx, agentID, limit)
}

func (s *AdminDashboardService) GetTransaction(ctx context.Context, txnID uuid.UUID) (*models.Transaction, error) {
	return s.txnRepo.GetByID(ctx, txnID)
}

func (s *AdminDashboardService) ListPendingTransactions(ctx context.Context, limit int) ([]models.Transaction, error) {
	return s.txnRepo.ListPending(ctx, limit)
}

func (s *AdminDashboardService) GetAgentHistory(ctx context.Context, agentID uuid.UUID, limit int) ([]models.Transaction, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	return s.txnRepo.ListByAgent(ctx, agentID, limit)
}

func (s *AdminDashboardService) ListApprovalAuditLogs(ctx context.Context, limit int) ([]models.ApprovalAuditLog, error) {
	return s.auditRepo.ListApprovalLogs(ctx, limit)
}

// ApprovePendingTransaction approves a pending transaction and deducts user balance atomically.
func (s *AdminDashboardService) ApprovePendingTransaction(ctx context.Context, txnID uuid.UUID, adminID uuid.UUID, requestID string) (*models.Transaction, error) {
	if adminID == uuid.Nil {
		return nil, fmt.Errorf("admin id is required")
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	txn, err := s.txnRepo.GetByIDForUpdate(ctx, tx, txnID)
	if err != nil {
		return nil, err
	}

	if txn.Status == "APPROVED" || txn.Status == "DENIED" {
		_ = tx.Commit()
		return txn, nil
	}
	if txn.Status != "PENDING_APPROVAL" {
		return nil, fmt.Errorf("transaction is not pending approval")
	}

	agent, err := s.agentRepo.GetByIDForUpdate(ctx, tx, txn.AgentID)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(agent.Status) == "frozen" {
		updated, uErr := s.txnRepo.UpdateStatus(ctx, tx, txn.ID, "DENIED", "agent_frozen")
		if uErr != nil {
			return nil, uErr
		}
		if err := s.auditRepo.CreateApprovalLog(ctx, tx, &models.CreateApprovalAuditLogRequest{
			TransactionID:  txn.ID,
			AdminID:        adminID,
			Action:         "approve",
			PreviousStatus: txn.Status,
			NewStatus:      updated.Status,
			Reason:         updated.Reason,
			RequestID:      requestID,
		}); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit denied transaction: %w", err)
		}
		return updated, nil
	}

	if _, err := s.userRepo.GetByIDForUpdate(ctx, tx, agent.UserID); err != nil {
		return nil, err
	}
	if err := s.userRepo.DeductBalance(ctx, tx, agent.UserID, txn.AmountCents); err != nil {
		if strings.Contains(err.Error(), "insufficient balance") {
			updated, uErr := s.txnRepo.UpdateStatus(ctx, tx, txn.ID, "DENIED", "insufficient_balance")
			if uErr != nil {
				return nil, uErr
			}
			if err := s.auditRepo.CreateApprovalLog(ctx, tx, &models.CreateApprovalAuditLogRequest{
				TransactionID:  txn.ID,
				AdminID:        adminID,
				Action:         "approve",
				PreviousStatus: txn.Status,
				NewStatus:      updated.Status,
				Reason:         updated.Reason,
				RequestID:      requestID,
			}); err != nil {
				return nil, err
			}
			if err := tx.Commit(); err != nil {
				return nil, fmt.Errorf("failed to commit denied transaction: %w", err)
			}
			return updated, nil
		}
		return nil, err
	}

	var (
		updated *models.Transaction
	)

	if s.payments.Enabled() {
		checkout, cerr := s.payments.CreateCheckoutSession(ctx, payments.CreateCheckoutRequest{
			TransactionID: txn.ID,
			AgentID:       agent.ID,
			AmountCents:   txn.AmountCents,
			Currency:      txn.Currency,
			Vendor:        txn.Vendor,
		})
		if cerr != nil {
			updated, err = s.txnRepo.UpdateStatusAndPayment(
				ctx, tx, txn.ID,
				"DENIED", "checkout_creation_failed",
				s.payments.Name(), "", "", "checkout_creation_failed", "",
			)
			if err != nil {
				return nil, err
			}
			if err := s.auditRepo.CreateApprovalLog(ctx, tx, &models.CreateApprovalAuditLogRequest{
				TransactionID:  txn.ID,
				AdminID:        adminID,
				Action:         "approve",
				PreviousStatus: txn.Status,
				NewStatus:      updated.Status,
				Reason:         updated.Reason,
				RequestID:      requestID,
			}); err != nil {
				return nil, err
			}
			if err := tx.Commit(); err != nil {
				return nil, fmt.Errorf("failed to commit denied transaction: %w", err)
			}
			return updated, nil
		}

		updated, err = s.txnRepo.UpdateStatusAndPayment(
			ctx, tx, txn.ID,
			"APPROVED", "approved",
			checkout.Provider, checkout.SessionID, checkout.PaymentIntentID, checkout.ProviderStatus, checkout.CheckoutURL,
		)
	} else {
		updated, err = s.txnRepo.UpdateStatusAndPayment(
			ctx, tx, txn.ID,
			"APPROVED", "approved",
			"", "", "", "not_applicable", "",
		)
	}
	if err != nil {
		return nil, err
	}

	if err := s.auditRepo.CreateApprovalLog(ctx, tx, &models.CreateApprovalAuditLogRequest{
		TransactionID:  txn.ID,
		AdminID:        adminID,
		Action:         "approve",
		PreviousStatus: txn.Status,
		NewStatus:      updated.Status,
		Reason:         updated.Reason,
		RequestID:      requestID,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit approved transaction: %w", err)
	}

	return updated, nil
}

// DenyPendingTransaction denies a pending transaction.
func (s *AdminDashboardService) DenyPendingTransaction(ctx context.Context, txnID uuid.UUID, adminID uuid.UUID, requestID string) (*models.Transaction, error) {
	if adminID == uuid.Nil {
		return nil, fmt.Errorf("admin id is required")
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	txn, err := s.txnRepo.GetByIDForUpdate(ctx, tx, txnID)
	if err != nil {
		return nil, err
	}

	if txn.Status == "DENIED" || txn.Status == "APPROVED" {
		_ = tx.Commit()
		return txn, nil
	}
	if txn.Status != "PENDING_APPROVAL" {
		return nil, fmt.Errorf("transaction is not pending approval")
	}

	updated, err := s.txnRepo.UpdateStatusAndPayment(
		ctx, tx, txn.ID,
		"DENIED", "human_denied",
		txn.Provider, txn.ProviderSessionID, txn.ProviderPaymentIntentID, txn.ProviderStatus, txn.ProviderCheckoutURL,
	)
	if err != nil {
		return nil, err
	}

	if err := s.auditRepo.CreateApprovalLog(ctx, tx, &models.CreateApprovalAuditLogRequest{
		TransactionID:  txn.ID,
		AdminID:        adminID,
		Action:         "deny",
		PreviousStatus: txn.Status,
		NewStatus:      updated.Status,
		Reason:         updated.Reason,
		RequestID:      requestID,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit denied transaction: %w", err)
	}

	return updated, nil
}
