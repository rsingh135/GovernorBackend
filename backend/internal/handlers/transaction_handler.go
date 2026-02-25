package handlers

import (
	"net/http"

	"agentpay/internal/httpjson"
	"agentpay/internal/httputil"
	"agentpay/internal/middleware"
	"agentpay/internal/models"
	"agentpay/internal/services"
)

// TransactionHandler handles transaction listing endpoints.
type TransactionHandler struct {
	txnService *services.TransactionService
}

// NewTransactionHandler creates a new transaction handler.
func NewTransactionHandler(txnService *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		txnService: txnService,
	}
}

// ListTransactions handles GET /transactions with filters and pagination.
func (h *TransactionHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	// Get authenticated agent from context
	agent, ok := middleware.GetAgentFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse filters
	filters := models.TransactionFilters{}

	// Filter by status if provided
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		// Convert lowercase API status to uppercase DB status
		upperStatus := statusStr
		if statusStr == "approved" {
			upperStatus = "APPROVED"
		} else if statusStr == "denied" {
			upperStatus = "DENIED"
		} else if statusStr == "pending_approval" {
			upperStatus = "PENDING_APPROVAL"
		}
		filters.Status = &upperStatus
	}

	// Filter by date range
	if fromDate, err := httputil.ParseTime(r, "from_date"); err == nil && fromDate != nil {
		filters.FromDate = fromDate
	}
	if toDate, err := httputil.ParseTime(r, "to_date"); err == nil && toDate != nil {
		filters.ToDate = toDate
	}

	// Enforce auth scoping: agent can only see own transactions
	filters.AgentID = &agent.ID

	// Parse pagination
	pagination := httputil.ParsePagination(r)

	// Call service
	resp, err := h.txnService.ListTransactions(r.Context(), filters, pagination)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpjson.Write(w, http.StatusOK, resp)
}
