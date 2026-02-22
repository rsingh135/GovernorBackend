package handlers

import (
	"net/http"
	"strings"

	"agentpay/internal/httpjson"
	"agentpay/internal/middleware"
	"agentpay/internal/models"
	"agentpay/internal/services"
)

// AdminAuthHandler handles admin auth endpoints for the dashboard.
type AdminAuthHandler struct {
	adminAuthService *services.AdminAuthService
}

// NewAdminAuthHandler creates a new admin auth handler.
func NewAdminAuthHandler(adminAuthService *services.AdminAuthService) *AdminAuthHandler {
	return &AdminAuthHandler{adminAuthService: adminAuthService}
}

// Login handles POST /admin/login.
func (h *AdminAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.AdminLoginRequest
	if ok := httpjson.DecodeStrict(w, r, &req); !ok {
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)
	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	resp, err := h.adminAuthService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	httpjson.Write(w, http.StatusOK, resp)
}

// Me handles GET /admin/me.
func (h *AdminAuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.GetAdminFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	httpjson.Write(w, http.StatusOK, models.AdminMeResponse{Admin: *admin})
}
