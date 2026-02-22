# Governor Backend

Governed spending API for AI agents. This repository currently contains the Go backend + Postgres migrations.

## Current Status

Implemented API routes:
- `POST /users`
- `POST /agents`
- `POST /policies`
- `POST /spend` (agent-authenticated)
- `POST /admin/login`
- `GET /admin/me` (admin-authenticated)
- `GET /admin/users`, `GET /admin/users/{id}` (admin-authenticated)
- `GET /admin/agents`, `GET /admin/agents/{id}`, `GET /admin/agents/{id}/history?limit=10` (admin-authenticated)
- `GET /admin/policies?agent_id={id}` (admin-authenticated)
- `GET /admin/transactions`, `GET /admin/transactions/{id}` (admin-authenticated)
- `GET /admin/transactions/pending` (admin-authenticated)
- `POST /admin/transactions/{id}/approve`, `POST /admin/transactions/{id}/deny` (admin-authenticated)
- `POST /webhooks/stripe`
- `GET /health`

## Prerequisites

- Go 1.21+
- Docker + Docker Compose

## Local Setup

1. Start Postgres:
```bash
make db-up
```

2. Apply incremental migrations:
```bash
make migrate
```

3. Set env vars (or use defaults shown in `.env.example`):
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=agentpay
export PORT=8080
export ADMIN_SESSION_TTL_HOURS=24
export STRIPE_SECRET_KEY=
export STRIPE_WEBHOOK_SECRET=
export STRIPE_SUCCESS_URL=http://localhost:3000/checkout/success?session_id={CHECKOUT_SESSION_ID}
export STRIPE_CANCEL_URL=http://localhost:3000/checkout/cancel
```

4. Start API:
```bash
make backend
```

## Test Setup

```bash
make db-test-setup
make test
```

Note: test execution requires Go installed on your machine.

## Seeded Admin (MVP Scaffold)

Migration `003_add_admin_auth.sql` seeds:
- Email: `admin@governor.local`
- Password: `governor_admin_123`

Login example:
```bash
curl -X POST http://localhost:8080/admin/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@governor.local","password":"governor_admin_123"}'
```

Then call `/admin/me`:
```bash
curl http://localhost:8080/admin/me \
  -H "Authorization: Bearer <TOKEN>"
```

## Agent Auth for Spend

`POST /spend` uses `apiKey` (or `X-API-Key`) header.

```bash
curl -X POST http://localhost:8080/spend \
  -H "Content-Type: application/json" \
  -H "apiKey: sk_test_agent_123" \
  -d '{
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "amount": 500,
    "vendor": "openai.com",
    "meta": {}
  }'
```

If `STRIPE_SECRET_KEY` is set, approved transactions include `checkout_url` and `provider_status`.

## Stripe Webhook (Local)

Forward Stripe events to local API:
```bash
stripe listen --forward-to localhost:8080/webhooks/stripe
```
