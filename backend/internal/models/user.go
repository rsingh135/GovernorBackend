package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents an account holder with a balance.
type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	BalanceCents int64     `json:"balance_cents"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateUserRequest is the request body for creating a user.
type CreateUserRequest struct {
	Name               string `json:"name"`
	InitialBalanceCents int64  `json:"initial_balance_cents"`
}

// CreateUserResponse is the response after user creation.
type CreateUserResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	BalanceCents int64     `json:"balance_cents"`
	CreatedAt    time.Time `json:"created_at"`
}
