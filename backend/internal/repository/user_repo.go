package repository

import (
	"context"
	"database/sql"
	"fmt"

	"agentpay/internal/models"

	"github.com/google/uuid"
)

// UserRepository handles user data persistence.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user with initial balance.
func (r *UserRepository) Create(ctx context.Context, name string, initialBalanceCents int64) (*models.User, error) {
	user := &models.User{}

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (name, balance_cents)
		VALUES ($1, $2)
		RETURNING id, name, balance_cents, created_at, updated_at
	`, name, initialBalanceCents).Scan(
		&user.ID,
		&user.Name,
		&user.BalanceCents,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetByID retrieves a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user := &models.User{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, balance_cents, created_at, updated_at
		FROM users
		WHERE id = $1
	`, userID).Scan(
		&user.ID,
		&user.Name,
		&user.BalanceCents,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByIDForUpdate retrieves a user with row-level lock (for balance updates).
// Must be called within a transaction.
func (r *UserRepository) GetByIDForUpdate(ctx context.Context, tx *sql.Tx, userID uuid.UUID) (*models.User, error) {
	user := &models.User{}

	err := tx.QueryRowContext(ctx, `
		SELECT id, name, balance_cents, created_at, updated_at
		FROM users
		WHERE id = $1
		FOR UPDATE
	`, userID).Scan(
		&user.ID,
		&user.Name,
		&user.BalanceCents,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to lock user: %w", err)
	}

	return user, nil
}

// DeductBalance deducts an amount from user balance.
// Must be called within a transaction after GetByIDForUpdate.
func (r *UserRepository) DeductBalance(ctx context.Context, tx *sql.Tx, userID uuid.UUID, amountCents int64) error {
	result, err := tx.ExecContext(ctx, `
		UPDATE users
		SET balance_cents = balance_cents - $1,
		    updated_at = now()
		WHERE id = $2 AND balance_cents >= $1
	`, amountCents, userID)

	if err != nil {
		return fmt.Errorf("failed to deduct balance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("insufficient balance")
	}

	return nil
}

// List retrieves users ordered by creation time descending.
func (r *UserRepository) List(ctx context.Context, limit int) ([]models.User, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, balance_cents, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := make([]models.User, 0, limit)
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.BalanceCents, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterating users: %w", err)
	}

	return users, nil
}
