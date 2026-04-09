package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"agentpay/internal/models"
	"agentpay/internal/services"

	"github.com/google/uuid"
)

// ApprovalHandler handles approve/deny HTTP requests.
type ApprovalHandler struct {
	service *services.ApprovalService
}

// NewApprovalHandler creates a new approval handler.
func NewApprovalHandler(service *services.ApprovalService) *ApprovalHandler {
	return &ApprovalHandler{service: service}
}

// Approve handles POST /transactions/:id/approve.
func (h *ApprovalHandler) Approve(w http.ResponseWriter, r *http.Request) {
	h.handleDecision(w, r, true)
}

// Deny handles POST /transactions/:id/deny.
func (h *ApprovalHandler) Deny(w http.ResponseWriter, r *http.Request) {
	h.handleDecision(w, r, false)
}

func (h *ApprovalHandler) handleDecision(w http.ResponseWriter, r *http.Request, approve bool) {
	// Extract transaction ID from path: /transactions/{id}/approve or /transactions/{id}/deny
	path := strings.TrimPrefix(r.URL.Path, "/transactions/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		http.Error(w, `{"error":"invalid path"}`, http.StatusBadRequest)
		return
	}

	txnID, err := uuid.Parse(parts[0])
	if err != nil {
		http.Error(w, `{"error":"invalid transaction ID"}`, http.StatusBadRequest)
		return
	}

	var req models.ApproveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.ApproverUserID == uuid.Nil {
		http.Error(w, `{"error":"approver_user_id is required"}`, http.StatusBadRequest)
		return
	}

	var resp *models.ApproveResponse
	if approve {
		resp, err = h.service.Approve(r.Context(), txnID, req.ApproverUserID)
	} else {
		resp, err = h.service.Deny(r.Context(), txnID, req.ApproverUserID)
	}

	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
