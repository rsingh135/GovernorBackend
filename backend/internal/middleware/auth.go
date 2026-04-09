package middleware

import (
	"context"
	"net/http"
	"strings"

	"agentpay/internal/models"
	"agentpay/internal/services"
)

type contextKey string

const AgentContextKey contextKey = "agent"

// AuthMiddleware authenticates requests using API key.
type AuthMiddleware struct {
	agentService *services.AgentService
}

// NewAuthMiddleware creates a new auth middleware.
func NewAuthMiddleware(agentService *services.AgentService) *AuthMiddleware {
	return &AuthMiddleware{
		agentService: agentService,
	}
}

// Authenticate validates the API key and adds agent to context.
func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract API key from header
		apiKey := strings.TrimSpace(r.Header.Get("apiKey"))
		if apiKey == "" {
			apiKey = strings.TrimSpace(r.Header.Get("X-API-Key"))
		}

		if apiKey == "" {
			http.Error(w, "missing api key", http.StatusUnauthorized)
			return
		}

		// Authenticate
		agent, err := m.agentService.AuthenticateAgent(r.Context(), apiKey)
		if err != nil {
			http.Error(w, "invalid api key", http.StatusUnauthorized)
			return
		}

		if agent.Status == "frozen" {
			http.Error(w, "agent frozen", http.StatusForbidden)
			return
		}

		// Add agent to context
		ctx := context.WithValue(r.Context(), AgentContextKey, agent)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetAgentFromContext retrieves the agent from request context.
func GetAgentFromContext(ctx context.Context) (*models.Agent, bool) {
	agent, ok := ctx.Value(AgentContextKey).(*models.Agent)
	return agent, ok
}
