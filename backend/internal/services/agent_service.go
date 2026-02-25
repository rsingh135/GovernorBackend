package services

import (
	"context"
	"database/sql"

	"agentpay/internal/models"
	"agentpay/internal/repository"

	"github.com/google/uuid"
)

// AgentService handles agent business logic.
type AgentService struct {
	agentRepo *repository.AgentRepository
	userRepo  *repository.UserRepository
}

// NewAgentService creates a new agent service.
func NewAgentService(db *sql.DB) *AgentService {
	return &AgentService{
		agentRepo: repository.NewAgentRepository(db),
		userRepo:  repository.NewUserRepository(db),
	}
}

// CreateAgent provisions a new agent for a user.
func (s *AgentService) CreateAgent(ctx context.Context, userID uuid.UUID, name string) (*models.CreateAgentResponse, error) {
	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Create agent
	agent, apiKey, err := s.agentRepo.Create(ctx, userID, name)
	if err != nil {
		return nil, err
	}

	return &models.CreateAgentResponse{
		ID:        agent.ID,
		UserID:    agent.UserID,
		Name:      agent.Name,
		Status:    agent.Status,
		APIKey:    apiKey,
		CreatedAt: agent.CreatedAt,
	}, nil
}

// AuthenticateAgent validates an API key and returns the agent.
func (s *AgentService) AuthenticateAgent(ctx context.Context, apiKey string) (*models.Agent, error) {
	return s.agentRepo.GetByAPIKey(ctx, apiKey)
}

// ListAgents retrieves paginated agents with filters.
func (s *AgentService) ListAgents(
	ctx context.Context,
	filters models.AgentFilters,
	pagination models.PaginationParams,
) (*models.ListAgentsResponse, error) {
	// Get agents
	agents, err := s.agentRepo.List(ctx, filters, pagination)
	if err != nil {
		return nil, err
	}

	// Get total count
	total, err := s.agentRepo.Count(ctx, filters)
	if err != nil {
		return nil, err
	}

	return &models.ListAgentsResponse{
		Agents: agents,
		PaginatedResponse: models.PaginatedResponse{
			Total:  total,
			Limit:  pagination.Limit,
			Offset: pagination.Offset,
		},
	}, nil
}
