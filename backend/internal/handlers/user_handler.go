package handlers

import (
	"net/http"

	"agentpay/internal/httpjson"
	"agentpay/internal/models"
	"agentpay/internal/services"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new user handler.
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUser handles POST /users.
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if ok := httpjson.DecodeStrict(w, r, &req); !ok {
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if req.InitialBalanceCents < 0 {
		http.Error(w, "initial balance must be non-negative", http.StatusBadRequest)
		return
	}

	resp, err := h.userService.CreateUser(r.Context(), req.Name, req.InitialBalanceCents)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpjson.Write(w, http.StatusCreated, resp)
}
