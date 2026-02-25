package services

import (
	"context"
	"database/sql"

	"agentpay/internal/models"
	"agentpay/internal/repository"

	"github.com/google/uuid"
)

// UserService handles user business logic.
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService creates a new user service.
func NewUserService(db *sql.DB) *UserService {
	return &UserService{
		userRepo: repository.NewUserRepository(db),
	}
}

// CreateUser creates a new user with initial balance.
func (s *UserService) CreateUser(ctx context.Context, name string, initialBalanceCents int64) (*models.CreateUserResponse, error) {
	user, err := s.userRepo.Create(ctx, name, initialBalanceCents)
	if err != nil {
		return nil, err
	}

	return &models.CreateUserResponse{
		ID:           user.ID,
		Name:         user.Name,
		BalanceCents: user.BalanceCents,
		CreatedAt:    user.CreatedAt,
	}, nil
}

// GetUser retrieves a user by ID.
func (s *UserService) GetUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}
