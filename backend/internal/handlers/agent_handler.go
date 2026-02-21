package handlers

import (
	"net/http"

	"agentpay/internal/httpjson"
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
