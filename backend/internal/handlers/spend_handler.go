package handlers

import (
	"net/http"

	"agentpay/internal/httpjson"
	"agentpay/internal/middleware"
	"agentpay/internal/models"
	"agentpay/internal/services"

	"github.com/google/uuid"
)

// SpendHandler handles spend-related HTTP requests.
type SpendHandler struct {
	spendService *services.SpendService
}

// NewSpendHandler creates a new spend handler.
func NewSpendHandler(spendService *services.SpendService) *SpendHandler {
	return &SpendHandler{
		spendService: spendService,
	}
}

// Spend handles POST /spend.
// CRITICAL: This endpoint enforces idempotency and uses row-level locking.
func (h *SpendHandler) Spend(w http.ResponseWriter, r *http.Request) {
	// Get authenticated agent from context (set by auth middleware)
	agent, ok := middleware.GetAgentFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.SpendRequest
	if ok := httpjson.DecodeStrict(w, r, &req); !ok {
		httpjson.Write(w, http.StatusBadRequest, models.SpendResponse{
			Status: "denied",
			Reason: "invalid_json",
		})
		return
	}

	// Validate request
	if req.RequestID == uuid.Nil {
		httpjson.Write(w, http.StatusBadRequest, models.SpendResponse{
			Status: "denied",
			Reason: "invalid_request_id",
		})
		return
	}

	if req.Amount <= 0 {
		httpjson.Write(w, http.StatusBadRequest, models.SpendResponse{
			Status: "denied",
			Reason: "amount_must_be_positive",
		})
		return
	}

	if req.Vendor == "" {
		httpjson.Write(w, http.StatusBadRequest, models.SpendResponse{
			Status: "denied",
			Reason: "vendor_required",
		})
		return
	}

	// Process spend
	resp, err := h.spendService.ProcessSpend(r.Context(), agent, &req)
	if err != nil {
		// Log error for debugging
		w.Header().Set("X-Error-Detail", err.Error())
		httpjson.Write(w, http.StatusInternalServerError, models.SpendResponse{
			Status: "denied",
			Reason: "internal_error",
		})
		return
	}

	httpjson.Write(w, http.StatusOK, resp)
}
