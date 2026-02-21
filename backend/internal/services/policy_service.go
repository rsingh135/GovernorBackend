package services

import (
	"context"
	"database/sql"

	"agentpay/internal/models"
	"agentpay/internal/repository"
)

// PolicyService handles policy business logic.
type PolicyService struct {
	policyRepo *repository.PolicyRepository
	agentRepo  *repository.AgentRepository
}

// NewPolicyService creates a new policy service.
func NewPolicyService(db *sql.DB) *PolicyService {
	return &PolicyService{
		policyRepo: repository.NewPolicyRepository(db),
		agentRepo:  repository.NewAgentRepository(db),
	}
}

// UpsertPolicy creates or updates a policy for an agent.
func (s *PolicyService) UpsertPolicy(ctx context.Context, req *models.UpsertPolicyRequest) (*models.Policy, error) {
	// Verify agent exists
	_, err := s.agentRepo.GetByID(ctx, req.AgentID)
	if err != nil {
		return nil, err
	}

	return s.policyRepo.Upsert(ctx, req)
}
