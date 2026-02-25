package handlers

import (
	"net/http"

	"agentpay/internal/httpjson"
	"agentpay/internal/models"
	"agentpay/internal/services"

	"github.com/google/uuid"
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

// GetUser handles GET /users/:id.
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path /users/:id
	idStr := r.URL.Path[len("/users/"):]
	if idStr == "" {
		http.Error(w, "user ID required", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetUser(r.Context(), id)
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpjson.Write(w, http.StatusOK, user)
}
