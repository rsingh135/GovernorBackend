# Governor Backend Quick Start

## 1. Start Database

```bash
make db-up
```

## 2. Apply Incremental Migrations

```bash
make migrate
```

## 3. Run Tests

```bash
make db-test-setup
make test
```

## 4. Start API

```bash
make backend
```

Server: `http://localhost:8080`

## 5. Verify Health

```bash
curl http://localhost:8080/health
```

## 6. Seeded Admin Login (MVP Scaffold)

```bash
curl -X POST http://localhost:8080/admin/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@governor.local","password":"governor_admin_123"}'
```

Use the returned token for admin dashboard APIs:
```bash
curl http://localhost:8080/admin/transactions/pending \
  -H "Authorization: Bearer <TOKEN>"
```

## 7. Optional Stripe Test Mode

Set Stripe env vars before `make backend`:
```bash
export STRIPE_SECRET_KEY=<YOUR_STRIPE_TEST_SECRET_KEY>
export STRIPE_WEBHOOK_SECRET=<YOUR_STRIPE_WEBHOOK_SECRET>
```

Forward webhooks:
```bash
stripe listen --forward-to localhost:8080/webhooks/stripe
```

## Available Make Commands

```bash
make db-up
make db-down
make db-reset
make migrate
make db-test-setup
make backend
make test
make test-verbose
```
