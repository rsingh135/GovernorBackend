package handlers

import (
	"net/http"
	"strings"

	"agentpay/internal/httpjson"
	"agentpay/internal/middleware"
	"agentpay/internal/models"
	"agentpay/internal/services"

	"github.com/google/uuid"
)

// WebhookHandler handles webhook registration endpoints.
type WebhookHandler struct {
	webhookService *services.WebhookService
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(webhookService *services.WebhookService) *WebhookHandler {
	return &WebhookHandler{webhookService: webhookService}
}

// Register handles POST /webhooks.
func (h *WebhookHandler) Register(w http.ResponseWriter, r *http.Request) {
	agent, ok := middleware.GetAgentFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.RegisterWebhookRequest
	if ok := httpjson.DecodeStrict(w, r, &req); !ok {
		return
	}

	resp, err := h.webhookService.Register(r.Context(), agent.ID, &req)
	if err != nil {
		code := webhookErrCode(err)
		http.Error(w, err.Error(), code)
		return
	}

	httpjson.Write(w, http.StatusCreated, resp)
}

// List handles GET /webhooks.
func (h *WebhookHandler) List(w http.ResponseWriter, r *http.Request) {
	agent, ok := middleware.GetAgentFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := h.webhookService.List(r.Context(), agent.ID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httpjson.Write(w, http.StatusOK, resp)
}

// Delete handles DELETE /webhooks/:id.
func (h *WebhookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	agent, ok := middleware.GetAgentFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/webhooks/")
	webhookID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid webhook ID", http.StatusBadRequest)
		return
	}

	if err := h.webhookService.Delete(r.Context(), webhookID, agent.ID); err != nil {
		code := webhookErrCode(err)
		http.Error(w, err.Error(), code)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// webhookErrCode maps service errors to HTTP status codes.
func webhookErrCode(err error) int {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "not found"):
		return http.StatusNotFound
	case strings.Contains(msg, "required"),
		strings.Contains(msg, "valid http"),
		strings.Contains(msg, "unknown event"),
		strings.Contains(msg, "at least 16"):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
