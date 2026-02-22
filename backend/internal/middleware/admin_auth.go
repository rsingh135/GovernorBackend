package middleware

import (
	"context"
	"net/http"
	"strings"

	"agentpay/internal/models"
	"agentpay/internal/services"
)

type adminContextKey string

const AdminContextKey adminContextKey = "admin"

// AdminAuthMiddleware authenticates dashboard admin requests using session tokens.
type AdminAuthMiddleware struct {
	adminAuthService *services.AdminAuthService
}

// NewAdminAuthMiddleware creates a new admin auth middleware.
func NewAdminAuthMiddleware(adminAuthService *services.AdminAuthService) *AdminAuthMiddleware {
	return &AdminAuthMiddleware{adminAuthService: adminAuthService}
}

// Authenticate validates bearer token and adds admin to context.
func (m *AdminAuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := strings.TrimSpace(r.Header.Get("Authorization"))
		if auth == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(auth, prefix) {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(auth, prefix))
		admin, err := m.adminAuthService.AuthenticateSession(r.Context(), token)
		if err != nil {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), AdminContextKey, admin)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetAdminFromContext retrieves the authenticated admin from request context.
func GetAdminFromContext(ctx context.Context) (*models.Admin, bool) {
	admin, ok := ctx.Value(AdminContextKey).(*models.Admin)
	return admin, ok
}
