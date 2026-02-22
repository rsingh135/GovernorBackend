package models

import (
	"time"

	"github.com/google/uuid"
)

// Admin is an operator account used for dashboard actions.
type Admin struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AdminLoginRequest is the request body for POST /admin/login.
type AdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AdminLoginResponse contains a session token for dashboard use.
type AdminLoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Admin     Admin     `json:"admin"`
}

// AdminMeResponse is the response body for GET /admin/me.
type AdminMeResponse struct {
	Admin Admin `json:"admin"`
}
