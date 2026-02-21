# 🚀 Governor API - Quick Start Guide

## Prerequisites
- Docker & Docker Compose installed
- Go 1.21+ installed

---

## 🏃 Get Started in 3 Commands

```bash
# 1. Start PostgreSQL database
make db-up

# 2. Run tests to verify everything works
make test

# 3. Start the API server
make backend
```

Server will start on `http://localhost:8080`

---

## 📋 All Available Commands

### Database Operations
```bash
make db-up          # Start PostgreSQL
make db-down        # Stop PostgreSQL
make db-reset       # Reset database (destroys all data)
make db-test-setup  # Create and migrate test database
```

### Development
```bash
make backend        # Run the API server
make test           # Run test suite
make test-verbose   # Run tests with verbose output
```

### Manual Operations
```bash
# Build the binary
cd backend
go build -o governor-api cmd/server/main.go

# Run the binary
./governor-api

# Run tests with coverage
go test ./internal/handlers -cover -v
```

---

## 🎯 Example API Workflow

### 1. Create a User
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice",
    "initial_balance_cents": 10000
  }'
```

Save the `id` from the response.

### 2. Create an Agent for that User
```bash
curl -X POST http://localhost:8080/agents \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "<USER_ID>",
    "name": "MyAgent"
  }'
```

**⚠️ Save the `api_key` - it's only shown once!**

### 3. Create a Policy for the Agent
```bash
curl -X POST http://localhost:8080/policies \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "<AGENT_ID>",
    "daily_limit_cents": 5000,
    "allowed_vendors": ["openai.com", "anthropic.com"],
    "require_approval_above_cents": 3000
  }'
```

### 4. Execute a Spend
```bash
curl -X POST http://localhost:8080/spend \
  -H "Content-Type: application/json" \
  -H "apiKey: <YOUR_API_KEY>" \
  -d '{
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "amount": 1500,
    "vendor": "openai.com",
    "meta": {"purpose": "test"}
  }'
```

**Response:**
```json
{
  "status": "approved",
  "reason": "approved"
}
```

---

## 🧪 Test Data (Already Seeded)

After running migrations, the database includes test data:

**User:**
- Name: Ranveer
- Balance: $50.00 (5000 cents)
- ID: `11111111-1111-1111-1111-111111111111`

**Agent:**
- Name: BhangraBot
- User: Ranveer
- API Key: `sk_test_agent_123`
- ID: `22222222-2222-2222-2222-222222222222`

**Policy:**
- Daily Limit: $20.00 (2000 cents)
- Allowed Vendors: `openai.com`
- Approval Threshold: $15.00 (1500 cents)

### Test with Seeded Data
```bash
curl -X POST http://localhost:8080/spend \
  -H "Content-Type: application/json" \
  -H "apiKey: sk_test_agent_123" \
  -d '{
    "request_id": "550e8400-e29b-41d4-a716-446655440001",
    "amount": 500,
    "vendor": "openai.com",
    "meta": {}
  }'
```

---

## 🐛 Troubleshooting

### Database Connection Failed
```bash
# Check if PostgreSQL is running
docker ps | grep agentpay-db

# Check logs
docker logs agentpay-db

# Restart database
make db-reset
```

### Test Failures
```bash
# Ensure test database exists
make db-test-setup

# Run with verbose output
make test-verbose
```

### Port Already in Use
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or use a different port
PORT=8081 make backend
```

---

## 📊 Monitoring

### Check Server Health
```bash
curl http://localhost:8080/health
```

### View Database State
```bash
# Connect to database
docker exec agentpay-db psql -U postgres -d agentpay

# View users
SELECT * FROM users;

# View agents
SELECT * FROM agents;

# View transactions
SELECT * FROM transactions ORDER BY created_at DESC LIMIT 10;

# Check today's spend for an agent
SELECT
    agent_id,
    COUNT(*) as transaction_count,
    SUM(amount_cents) as total_spent_cents,
    SUM(amount_cents) / 100.0 as total_spent_dollars
FROM transactions
WHERE status = 'APPROVED'
    AND created_at >= date_trunc('day', now())
GROUP BY agent_id;
```

---

## 🔑 Environment Variables

```bash
# Database configuration (optional, defaults shown)
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=agentpay

# Server configuration
export PORT=8080

# For testing
export DB_NAME_TEST=agentpay_test
```

---

## 📖 Documentation

- **Full Implementation Details**: See `PHASE_1_COMPLETE.md`
- **Testing Guide**: See `TESTING_GUIDE.md`
- **Architecture**: See `PHASE_1_COMPLETE.md` → Architecture Overview

---

## 🎯 Key Features

✅ **User account balances** - Track funds per user
✅ **Agent provisioning** - Secure API keys for AI agents
✅ **Spending policies** - Daily limits, vendor allowlists, approval thresholds
✅ **Idempotent transactions** - Safe retry with `request_id`
✅ **Concurrency-safe** - Row-level locking prevents race conditions
✅ **Integer currency** - No floating-point errors

---

**Ready to build financial middleware for AI agents! 🤖💰**
