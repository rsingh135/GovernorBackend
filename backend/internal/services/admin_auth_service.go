package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"time"

	"agentpay/internal/models"
	"agentpay/internal/repository"

	"database/sql"
)

// AdminAuthService handles admin login/session auth for dashboard endpoints.
type AdminAuthService struct {
	adminRepo   *repository.AdminRepository
	sessionTTL  time.Duration
}

// NewAdminAuthService creates a new admin auth service.
func NewAdminAuthService(db *sql.DB, sessionTTL time.Duration) *AdminAuthService {
	return &AdminAuthService{
		adminRepo:  repository.NewAdminRepository(db),
		sessionTTL: sessionTTL,
	}
}

// Login validates credentials and creates a new session token.
func (s *AdminAuthService) Login(ctx context.Context, email, password string) (*models.AdminLoginResponse, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	password = strings.TrimSpace(password)

	admin, err := s.adminRepo.GetByEmailAndPasswordHash(ctx, email, sha256Hex(password))
	if err != nil {
		return nil, err
	}

	token, err := randomToken(32)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(s.sessionTTL)
	if err := s.adminRepo.CreateSession(ctx, admin.ID, sha256Hex(token), expiresAt); err != nil {
		return nil, err
	}

	return &models.AdminLoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		Admin:     *admin,
	}, nil
}

// AuthenticateSession validates a session token and returns the matching admin.
func (s *AdminAuthService) AuthenticateSession(ctx context.Context, token string) (*models.Admin, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, sql.ErrNoRows
	}
	return s.adminRepo.GetBySessionTokenHash(ctx, sha256Hex(token))
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
