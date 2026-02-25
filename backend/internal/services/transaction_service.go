package services

import (
	"context"
	"database/sql"

	"agentpay/internal/models"
	"agentpay/internal/repository"
)

// TransactionService handles transaction business logic.
type TransactionService struct {
	db      *sql.DB
	txnRepo *repository.TransactionRepository
}

// NewTransactionService creates a new transaction service.
func NewTransactionService(db *sql.DB) *TransactionService {
	return &TransactionService{
		db:      db,
		txnRepo: repository.NewTransactionRepository(db),
	}
}

// ListTransactions retrieves paginated transactions with filters.
func (s *TransactionService) ListTransactions(
	ctx context.Context,
	filters models.TransactionFilters,
	pagination models.PaginationParams,
) (*models.ListTransactionsResponse, error) {
	// Get transactions
	transactions, err := s.txnRepo.List(ctx, filters, pagination)
	if err != nil {
		return nil, err
	}

	// Get total count
	total, err := s.txnRepo.Count(ctx, filters)
	if err != nil {
		return nil, err
	}

	// Convert uppercase DB status to lowercase API status
	for i := range transactions {
		if transactions[i].Status == "APPROVED" {
			transactions[i].Status = "approved"
		} else if transactions[i].Status == "DENIED" {
			transactions[i].Status = "denied"
		} else if transactions[i].Status == "PENDING_APPROVAL" {
			transactions[i].Status = "pending_approval"
		}
	}

	return &models.ListTransactionsResponse{
		Transactions: transactions,
		PaginatedResponse: models.PaginatedResponse{
			Total:  total,
			Limit:  pagination.Limit,
			Offset: pagination.Offset,
		},
	}, nil
}
