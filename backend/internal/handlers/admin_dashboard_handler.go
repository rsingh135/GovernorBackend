package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"agentpay/internal/httpjson"
	"agentpay/internal/middleware"
	"agentpay/internal/models"
	"agentpay/internal/services"

	"github.com/google/uuid"
)

// AdminDashboardHandler handles admin dashboard read/review endpoints.
type AdminDashboardHandler struct {
	service *services.AdminDashboardService
}

// NewAdminDashboardHandler creates a new admin dashboard handler.
func NewAdminDashboardHandler(service *services.AdminDashboardService) *AdminDashboardHandler {
	return &AdminDashboardHandler{service: service}
}

func (h *AdminDashboardHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.ListUsers(r.Context(), parseLimit(r, 50))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]interface{}{"users": users})
}

func (h *AdminDashboardHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUUIDPath(r.URL.Path, "/admin/users/")
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	user, err := h.service.GetUser(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]interface{}{"user": user})
}

func (h *AdminDashboardHandler) FreezeUser(w http.ResponseWriter, r *http.Request) {
	userID, action, err := parseEntityActionPath(r.URL.Path, "/admin/users/")
	if err != nil || action != "freeze" {
		http.Error(w, "invalid user action path", http.StatusBadRequest)
		return
	}
	user, err := h.service.FreezeUser(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]interface{}{"user": user})
}

func (h *AdminDashboardHandler) UnfreezeUser(w http.ResponseWriter, r *http.Request) {
	userID, action, err := parseEntityActionPath(r.URL.Path, "/admin/users/")
	if err != nil || action != "unfreeze" {
		http.Error(w, "invalid user action path", http.StatusBadRequest)
		return
	}
	user, err := h.service.UnfreezeUser(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]interface{}{"user": user})
}

func (h *AdminDashboardHandler) ListAgents(w http.ResponseWriter, r *http.Request) {
	var userID *uuid.UUID
	if raw := strings.TrimSpace(r.URL.Query().Get("user_id")); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			http.Error(w, "invalid user_id", http.StatusBadRequest)
			return
		}
		userID = &parsed
	}

	agents, err := h.service.ListAgents(r.Context(), userID, parseLimit(r, 50))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]interface{}{"agents": agents})
}

func (h *AdminDashboardHandler) GetAgent(w http.ResponseWriter, r *http.Request) {
	agentID, err := parseUUIDPath(r.URL.Path, "/admin/agents/")
	if err != nil {
		http.Error(w, "invalid agent id", http.StatusBadRequest)
		return
	}

	agent, err := h.service.GetAgent(r.Context(), agentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]interface{}{"agent": agent})
}

func (h *AdminDashboardHandler) FreezeAgent(w http.ResponseWriter, r *http.Request) {
	agentID, action, err := parseEntityActionPath(r.URL.Path, "/admin/agents/")
	if err != nil || action != "freeze" {
		http.Error(w, "invalid agent action path", http.StatusBadRequest)
		return
	}
	agent, err := h.service.FreezeAgent(r.Context(), agentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]interface{}{"agent": agent})
}

func (h *AdminDashboardHandler) UnfreezeAgent(w http.ResponseWriter, r *http.Request) {
	agentID, action, err := parseEntityActionPath(r.URL.Path, "/admin/agents/")
	if err != nil || action != "unfreeze" {
		http.Error(w, "invalid agent action path", http.StatusBadRequest)
		return
	}
	agent, err := h.service.UnfreezeAgent(r.Context(), agentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]interface{}{"agent": agent})
}

func (h *AdminDashboardHandler) GetPolicyByAgent(w http.ResponseWriter, r *http.Request) {
	agentRaw := strings.TrimSpace(r.URL.Query().Get("agent_id"))
	if agentRaw == "" {
		http.Error(w, "agent_id is required", http.StatusBadRequest)
		return
	}

	agentID, err := uuid.Parse(agentRaw)
	if err != nil {
		http.Error(w, "invalid agent_id", http.StatusBadRequest)
		return
	}

	policy, err := h.service.GetPolicyByAgent(r.Context(), agentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]interface{}{"policy": policy})
}

func (h *AdminDashboardHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	var agentID *uuid.UUID
	if raw := strings.TrimSpace(r.URL.Query().Get("agent_id")); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			http.Error(w, "invalid agent_id", http.StatusBadRequest)
			return
		}
		agentID = &parsed
	}

	txs, err := h.service.ListTransactions(r.Context(), agentID, parseLimit(r, 50))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]interface{}{"transactions": normalizeTransactions(txs)})
}

func (h *AdminDashboardHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	txnID, err := parseUUIDPath(r.URL.Path, "/admin/transactions/")
	if err != nil {
		http.Error(w, "invalid transaction id", http.StatusBadRequest)
		return
	}

	txn, err := h.service.GetTransaction(r.Context(), txnID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]interface{}{"transaction": normalizeTransaction(*txn)})
}

func (h *AdminDashboardHandler) ListPendingTransactions(w http.ResponseWriter, r *http.Request) {
	txs, err := h.service.ListPendingTransactions(r.Context(), parseLimit(r, 50))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]interface{}{"transactions": normalizeTransactions(txs)})
}

func (h *AdminDashboardHandler) GetAgentHistory(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin/agents/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 || parts[1] != "history" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	agentID, err := uuid.Parse(parts[0])
	if err != nil {
		http.Error(w, "invalid agent id", http.StatusBadRequest)
		return
	}

	history, err := h.service.GetAgentHistory(r.Context(), agentID, parseLimit(r, 10))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]interface{}{"transactions": normalizeTransactions(history)})
}

func (h *AdminDashboardHandler) ListApprovalAuditLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.service.ListApprovalAuditLogs(r.Context(), parseLimit(r, 50))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]interface{}{"logs": logs})
}

func (h *AdminDashboardHandler) ApproveTransaction(w http.ResponseWriter, r *http.Request) {
	txnID, action, err := parseTransactionActionPath(r.URL.Path)
	if err != nil || action != "approve" {
		http.Error(w, "invalid transaction action path", http.StatusBadRequest)
		return
	}

	admin, ok := middleware.GetAdminFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	txn, err := h.service.ApprovePendingTransaction(r.Context(), txnID, admin.ID, middleware.GetRequestID(r.Context()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]interface{}{"transaction": normalizeTransaction(*txn)})
}

func (h *AdminDashboardHandler) DenyTransaction(w http.ResponseWriter, r *http.Request) {
	txnID, action, err := parseTransactionActionPath(r.URL.Path)
	if err != nil || action != "deny" {
		http.Error(w, "invalid transaction action path", http.StatusBadRequest)
		return
	}

	admin, ok := middleware.GetAdminFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	txn, err := h.service.DenyPendingTransaction(r.Context(), txnID, admin.ID, middleware.GetRequestID(r.Context()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]interface{}{"transaction": normalizeTransaction(*txn)})
}

func parseLimit(r *http.Request, defaultValue int) int {
	limitRaw := strings.TrimSpace(r.URL.Query().Get("limit"))
	if limitRaw == "" {
		return defaultValue
	}
	limit, err := strconv.Atoi(limitRaw)
	if err != nil {
		return defaultValue
	}
	return limit
}

func parseUUIDPath(path string, prefix string) (uuid.UUID, error) {
	raw := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	return uuid.Parse(raw)
}

func parseTransactionActionPath(path string) (uuid.UUID, string, error) {
	path = strings.TrimPrefix(path, "/admin/transactions/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 {
		return uuid.Nil, "", fmt.Errorf("invalid transaction action path")
	}

	txnID, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, "", err
	}

	return txnID, parts[1], nil
}

func parseEntityActionPath(path string, prefix string) (uuid.UUID, string, error) {
	path = strings.TrimPrefix(path, prefix)
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 {
		return uuid.Nil, "", fmt.Errorf("invalid action path")
	}

	entityID, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, "", err
	}

	return entityID, parts[1], nil
}

func normalizeTransactions(txs []models.Transaction) []models.Transaction {
	out := make([]models.Transaction, 0, len(txs))
	for _, tx := range txs {
		out = append(out, normalizeTransaction(tx))
	}
	return out
}

func normalizeTransaction(tx models.Transaction) models.Transaction {
	tx.Status = strings.ToLower(tx.Status)
	return tx
}
