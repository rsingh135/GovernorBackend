package handlers

import (
	"net/http"

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
	// Parse filters
	filters := models.AgentFilters{}

	// Filter by user_id if provided
	if userID, err := httputil.ParseUUID(r, "user_id"); err == nil && userID != nil {
		filters.UserID = userID
	} else if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	// Filter by status if provided
	if statusStr := httputil.ParseString(r, "status"); statusStr != nil {
		filters.Status = statusStr
	}

	// Parse pagination
	pagination := httputil.ParsePagination(r)

	// Call service
	resp, err := h.agentService.ListAgents(r.Context(), filters, pagination)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpjson.Write(w, http.StatusOK, resp)
}
