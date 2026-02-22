package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"agentpay/internal/models"

	"github.com/google/uuid"
)

// AdminRepository handles admin auth persistence.
type AdminRepository struct {
	db *sql.DB
}

// NewAdminRepository creates a new admin repository.
func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// GetByEmailAndPasswordHash retrieves an admin by login credentials.
func (r *AdminRepository) GetByEmailAndPasswordHash(ctx context.Context, email, passwordHash string) (*models.Admin, error) {
	admin := &models.Admin{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, email, created_at, updated_at
		FROM admins
		WHERE email = $1 AND password_hash = $2
	`, email, passwordHash).Scan(&admin.ID, &admin.Email, &admin.CreatedAt, &admin.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid credentials")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}
	return admin, nil
}

// CreateSession stores a hashed session token for an admin.
func (r *AdminRepository) CreateSession(ctx context.Context, adminID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO admin_sessions (admin_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, adminID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create admin session: %w", err)
	}
	return nil
}

// GetBySessionTokenHash returns an admin for a valid non-expired session token.
func (r *AdminRepository) GetBySessionTokenHash(ctx context.Context, tokenHash string) (*models.Admin, error) {
	admin := &models.Admin{}
	err := r.db.QueryRowContext(ctx, `
		SELECT a.id, a.email, a.created_at, a.updated_at
		FROM admins a
		JOIN admin_sessions s ON s.admin_id = a.id
		WHERE s.token_hash = $1
		  AND s.expires_at > now()
	`, tokenHash).Scan(&admin.ID, &admin.Email, &admin.CreatedAt, &admin.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid session")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get admin by session: %w", err)
	}
	return admin, nil
}
