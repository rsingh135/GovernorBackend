package handlers

import (
	"net/http"
	"strings"

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

	if req.PerTransactionLimitCents <= 0 {
		req.PerTransactionLimitCents = req.DailyLimitCents
	}
	if req.PerTransactionLimitCents > req.DailyLimitCents {
		http.Error(w, "per_transaction_limit_cents cannot exceed daily_limit_cents", http.StatusBadRequest)
		return
	}

	if req.RequireApprovalAboveCents < 0 {
		http.Error(w, "require_approval_above_cents must be non-negative", http.StatusBadRequest)
		return
	}
	if req.RequireApprovalAboveCents > req.PerTransactionLimitCents {
		http.Error(w, "require_approval_above_cents cannot exceed per_transaction_limit_cents", http.StatusBadRequest)
		return
	}

	if len(req.AllowedWeekdaysUTC) == 0 {
		req.AllowedWeekdaysUTC = []int{0, 1, 2, 3, 4, 5, 6}
	}
	for _, day := range req.AllowedWeekdaysUTC {
		if day < 0 || day > 6 {
			http.Error(w, "allowed_weekdays_utc values must be between 0 and 6", http.StatusBadRequest)
			return
		}
	}
	if len(req.AllowedHoursUTC) == 0 {
		req.AllowedHoursUTC = []int{
			0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
			12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23,
		}
	}
	for _, hour := range req.AllowedHoursUTC {
		if hour < 0 || hour > 23 {
			http.Error(w, "allowed_hours_utc values must be between 0 and 23", http.StatusBadRequest)
			return
		}
	}

	req.PurchaseGuideline = strings.TrimSpace(req.PurchaseGuideline)
	req.AllowedVendors = normalizeStringList(req.AllowedVendors)
	req.AllowedMCCs = normalizeStringList(req.AllowedMCCs)

	resp, err := h.policyService.UpsertPolicy(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpjson.Write(w, http.StatusOK, resp)
}

func normalizeStringList(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, raw := range values {
		value := strings.ToLower(strings.TrimSpace(raw))
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
