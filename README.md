# AgentPay MVP

Programmable Wallets for AI - A middleware API that governs financial transactions for AI agents.

## Tech Stack

- **Backend**: Go (standard `net/http`, `database/sql`, `lib/pq`)
- **Frontend**: React (Vite, TypeScript, TailwindCSS)
- **Database**: PostgreSQL

## Core Principles

1. **NO FLOAT MATH**: All financial values are integers representing cents (e.g., $10.00 = 1000)
2. **Type Safety**: Strict typing across Go and TypeScript
3. **Simplicity**: Clean, modular code without over-engineering
4. **Security**: No hardcoded secrets

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose (for database)

### Setup

1. **Start the database:**
   ```bash
   docker-compose up -d
   ```

2. **Set up the backend:**
   ```bash
   cd backend
   go mod download
   ```

3. **Set up the frontend:**
   ```bash
   cd frontend
   npm install
   ```

4. **Run the backend:**
   ```bash
   cd backend
   export DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres DB_NAME=agentpay PORT=8080
   go run cmd/api/main.go
   ```

5. **Run the frontend:**
   ```bash
   cd frontend
   npm run dev
   ```

The frontend will be available at `http://localhost:3000` and the API at `http://localhost:8080`.

## API Endpoints

### Public Endpoints

- `GET /api/users` - List all users
- `GET /api/users/{id}` - Get user by ID
- `POST /api/agents` - Create a new agent
- `GET /api/agents?user_id={id}` - List agents for a user

### Protected Endpoints (Require API Key)

- `POST /api/transactions` - Create a transaction
  - Headers: `Authorization: Bearer {api_key}`
  - Body: `{ "amount_cents": 1000, "vendor": "Example Vendor" }`
- `GET /api/transactions` - List transactions for authenticated agent
  - Headers: `Authorization: Bearer {api_key}`

## Database Schema

- **users**: Human users with balances
- **agents**: AI agents with daily spending limits
- **transactions**: Financial transactions with approval/decline status

See `migrations/001_initial_schema.sql` for the full schema.

## Environment Variables

### Backend

- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database user (default: postgres)
- `DB_PASSWORD` - Database password (default: postgres)
- `DB_NAME` - Database name (default: agentpay)
- `PORT` - API server port (default: 8080)

### Frontend

- `VITE_API_URL` - API base URL (default: /api)

## Example Usage

### Create an Agent

```bash
curl -X POST http://localhost:8080/api/agents \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "11111111-1111-1111-1111-111111111111",
    "name": "MyAgent",
    "daily_limit_cents": 5000
  }'
```

### Create a Transaction

```bash
curl -X POST http://localhost:8080/api/transactions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk_test_agent_123" \
  -d '{
    "amount_cents": 500,
    "vendor": "Coffee Shop"
  }'
```

## Project Structure

```
GovernorApp/
├── backend/
│   ├── cmd/
│   │   └── api/
│   │       └── main.go
│   ├── internal/
│   │   ├── handlers/    # HTTP handlers
│   │   ├── models/      # Data models
│   │   ├── db/          # Database connection
│   │   └── middleware/  # Auth middleware
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── components/  # React components
│   │   ├── pages/       # Page components
│   │   ├── types/       # TypeScript types
│   │   └── utils/       # Utility functions
│   └── package.json
├── migrations/
│   └── 001_initial_schema.sql
└── docker-compose.yml
```
