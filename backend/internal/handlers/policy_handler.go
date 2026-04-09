package handlers

import (
	"net/http"

	"agentpay/internal/httpjson"
	"agentpay/internal/models"
	"agentpay/internal/services"

	"github.com/google/uuid"
)

// PolicyHandler handles policy-related HTTP requests.
type PolicyHandler struct {
	policyService *services.PolicyService
}

// NewPolicyHandler creates a new policy handler.
func NewPolicyHandler(policyService *services.PolicyService) *PolicyHandler {
	return &PolicyHandler{
		policyService: policyService,
	}
}

// UpsertPolicy handles POST /policies.
func (h *PolicyHandler) UpsertPolicy(w http.ResponseWriter, r *http.Request) {
	var req models.UpsertPolicyRequest
	if ok := httpjson.DecodeStrict(w, r, &req); !ok {
		return
	}

	if req.AgentID == uuid.Nil {
		http.Error(w, "agent_id is required", http.StatusBadRequest)
		return
	}

	if req.DailyLimitCents <= 0 {
		http.Error(w, "daily_limit_cents must be positive", http.StatusBadRequest)
		return
	}

	if req.RequireApprovalAboveCents < 0 {
		http.Error(w, "require_approval_above_cents must be non-negative", http.StatusBadRequest)
		return
	}

	resp, err := h.policyService.UpsertPolicy(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpjson.Write(w, http.StatusOK, resp)
}

// GetPolicy handles GET /policies/:agent_id.
func (h *PolicyHandler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	// Extract agent_id from path /policies/:agent_id
	agentIDStr := r.URL.Path[len("/policies/"):]
	if agentIDStr == "" {
		http.Error(w, "agent_id required", http.StatusBadRequest)
		return
	}

	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		http.Error(w, "invalid agent_id", http.StatusBadRequest)
		return
	}

	policy, err := h.policyService.GetPolicyByAgentID(r.Context(), agentID)
	if err != nil {
		if err.Error() == "policy not found" {
			http.Error(w, "policy not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpjson.Write(w, http.StatusOK, policy)
}
