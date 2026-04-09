package handlers

import (
	"net/http"
	"strings"

	"agentpay/internal/httpjson"
	"agentpay/internal/httputil"
	"agentpay/internal/models"
	"agentpay/internal/services"

	"github.com/google/uuid"
)

// AgentHandler handles agent-related HTTP requests.
type AgentHandler struct {
	agentService *services.AgentService
}

// NewAgentHandler creates a new agent handler.
func NewAgentHandler(agentService *services.AgentService) *AgentHandler {
	return &AgentHandler{
		agentService: agentService,
	}
}

// CreateAgent handles POST /agents.
func (h *AgentHandler) CreateAgent(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAgentRequest
	if ok := httpjson.DecodeStrict(w, r, &req); !ok {
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if req.UserID == uuid.Nil {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	resp, err := h.agentService.CreateAgent(r.Context(), req.UserID, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpjson.Write(w, http.StatusCreated, resp)
}

// ListAgents handles GET /agents with filters and pagination.
func (h *AgentHandler) ListAgents(w http.ResponseWriter, r *http.Request) {
	filters := models.AgentFilters{}

	if userID, err := httputil.ParseUUID(r, "user_id"); err == nil && userID != nil {
		filters.UserID = userID
	} else if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	if statusStr := httputil.ParseString(r, "status"); statusStr != nil {
		filters.Status = statusStr
	}

	pagination := httputil.ParsePagination(r)

	resp, err := h.agentService.ListAgents(r.Context(), filters, pagination)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpjson.Write(w, http.StatusOK, resp)
}

// UpdateStatusRequest is the request body for PATCH /agents/:id/status.
type UpdateStatusRequest struct {
	Status string `json:"status"`
}

// UpdateStatus handles PATCH /agents/:id/status.
func (h *AgentHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	// Extract agent ID from path: /agents/{id}/status
	path := strings.TrimPrefix(r.URL.Path, "/agents/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "status" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	agentID, err := uuid.Parse(parts[0])
	if err != nil {
		http.Error(w, "invalid agent ID", http.StatusBadRequest)
		return
	}

	var req UpdateStatusRequest
	if ok := httpjson.DecodeStrict(w, r, &req); !ok {
		return
	}

	if req.Status != "active" && req.Status != "frozen" {
		http.Error(w, "status must be 'active' or 'frozen'", http.StatusBadRequest)
		return
	}

	err = h.agentService.SetStatus(r.Context(), agentID, req.Status)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "agent not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpjson.Write(w, http.StatusOK, map[string]string{
		"id":     agentID.String(),
		"status": req.Status,
	})
}
