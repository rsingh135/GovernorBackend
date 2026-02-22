package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"agentpay/internal/apikey"
	"agentpay/internal/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// SetupTestDB creates a test database connection and returns it.
// The caller is responsible for calling cleanup when done.
func SetupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	// Use test database or fallback to main DB with _test suffix
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME_TEST", "agentpay_test")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	// Clean all tables before test
	cleanup := func() {
		_, _ = db.Exec("TRUNCATE TABLE approval_audit_logs CASCADE")
		_, _ = db.Exec("TRUNCATE TABLE payment_webhook_events CASCADE")
		_, _ = db.Exec("TRUNCATE TABLE admin_sessions CASCADE")
		_, _ = db.Exec("TRUNCATE TABLE transactions CASCADE")
		_, _ = db.Exec("TRUNCATE TABLE policies CASCADE")
		_, _ = db.Exec("TRUNCATE TABLE agents CASCADE")
		_, _ = db.Exec("TRUNCATE TABLE users CASCADE")
		db.Close()
	}

	// Clean tables before starting
	_, _ = db.Exec("TRUNCATE TABLE approval_audit_logs CASCADE")
	_, _ = db.Exec("TRUNCATE TABLE payment_webhook_events CASCADE")
	_, _ = db.Exec("TRUNCATE TABLE admin_sessions CASCADE")
	_, _ = db.Exec("TRUNCATE TABLE transactions CASCADE")
	_, _ = db.Exec("TRUNCATE TABLE policies CASCADE")
	_, _ = db.Exec("TRUNCATE TABLE agents CASCADE")
	_, _ = db.Exec("TRUNCATE TABLE users CASCADE")

	return db, cleanup
}

// SeedTestData seeds the database with test users and agents.
func SeedTestData(t *testing.T, db *sql.DB) (ranveerUserID, bhangraBotAgentID uuid.UUID, bhangraBotAPIKey string) {
	t.Helper()
	ctx := context.Background()

	// Create Ranveer user with 5000 cents ($50) balance
	ranveerUserID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, name, balance_cents, created_at, updated_at)
		VALUES ($1, $2, $3, now(), now())
	`, ranveerUserID, "Ranveer", 5000)
	if err != nil {
		t.Fatalf("Failed to seed Ranveer user: %v", err)
	}

	// Create TigerBot agent for Ranveer with API key
	bhangraBotAgentID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	bhangraBotAPIKey = "sk_test_bhangra_bot_12345"
	keyHash := apikey.Hash(bhangraBotAPIKey)

	_, err = db.ExecContext(ctx, `
		INSERT INTO agents (id, user_id, name, status, api_key_hash, api_key_prefix, created_at)
		VALUES ($1, $2, $3, 'active', $4, 'sk_test_bhangr', now())
	`, bhangraBotAgentID, ranveerUserID, "TigerBot", keyHash)
	if err != nil {
		t.Fatalf("Failed to seed TigerBot agent: %v", err)
	}

	// Create policy for TigerBot: $100 daily limit, only openai.com allowed
	_, err = db.ExecContext(ctx, `
		INSERT INTO policies (agent_id, daily_limit_cents, allowed_vendors, require_approval_above_cents, raw_policy, created_at, updated_at)
		VALUES ($1, $2, $3, $4, '{}'::jsonb, now(), now())
	`, bhangraBotAgentID, 10000, pq.Array([]string{"openai.com"}), 1500)
	if err != nil {
		t.Fatalf("Failed to seed TigerBot policy: %v", err)
	}

	return ranveerUserID, bhangraBotAgentID, bhangraBotAPIKey
}

// GetUserBalance retrieves the current balance for a user.
func GetUserBalance(t *testing.T, db *sql.DB, userID uuid.UUID) int64 {
	t.Helper()

	var balance int64
	err := db.QueryRow("SELECT balance_cents FROM users WHERE id = $1", userID).Scan(&balance)
	if err != nil {
		t.Fatalf("Failed to get user balance: %v", err)
	}
	return balance
}

// CreateTestTransaction creates a test transaction directly in DB.
func CreateTestTransaction(t *testing.T, db *sql.DB, txn *models.Transaction) {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO transactions (id, request_id, agent_id, amount_cents, currency, vendor, status, reason, meta, created_at)
		VALUES ($1, $2, $3, $4, 'usd', $5, $6, $7, '{}'::jsonb, now())
	`, uuid.New(), txn.RequestID, txn.AgentID, txn.AmountCents, txn.Vendor, txn.Status, txn.Reason)

	if err != nil {
		t.Fatalf("Failed to create test transaction: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
