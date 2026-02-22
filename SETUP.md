# Governor Backend Setup

## Prerequisites

- Go 1.21+
- Docker + Docker Compose

## Install Go (macOS)

Homebrew:
```bash
brew install go
```

Verify:
```bash
go version
```

## Project Setup

1. Start DB:
```bash
make db-up
```

2. Apply incremental migrations:
```bash
make migrate
```

3. Run API:
```bash
make backend
```

## Default Environment

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=agentpay
PORT=8080
ADMIN_SESSION_TTL_HOURS=24
STRIPE_SECRET_KEY=
STRIPE_WEBHOOK_SECRET=
STRIPE_SUCCESS_URL=http://localhost:3000/checkout/success?session_id={CHECKOUT_SESSION_ID}
STRIPE_CANCEL_URL=http://localhost:3000/checkout/cancel
STRIPE_BASE_URL=
```

## Seeded Admin (MVP scaffold)

- Email: `admin@governor.local`
- Password: `governor_admin_123`

## Stripe Local Webhooks (Optional)

```bash
stripe listen --forward-to localhost:8080/webhooks/stripe
```

## Troubleshooting

- If `go` is not found, install Go and restart your terminal.
- If DB is stale, run:
```bash
make db-reset
make migrate
```
