# Governor Backend Testing Guide

## Prerequisites

- Go 1.21+
- Docker + Docker Compose

## 1. Start Database

```bash
make db-up
make migrate
```

## 2. Prepare Test Database

```bash
make db-test-setup
```

This applies:
- `001_initial_schema.sql`
- `002_add_users_table.sql`
- `003_add_admin_auth.sql`
- `004_add_payment_provider_fields.sql`
- `005_add_approval_audit_logs.sql`

## 3. Run Tests

```bash
make test
```

Verbose:
```bash
make test-verbose
```

Optional frontend type/build check:
```bash
make frontend-build
```

## 4. Run a Single Test (manual)

```bash
cd backend
DB_NAME_TEST=agentpay_test go test ./internal/handlers -run TestSpendHandler_Idempotency -v
```

## Notes

- Current integration tests cover spend flow behavior and idempotency in `backend/internal/handlers/api_test.go`.
- Admin dashboard review flow is covered in `backend/internal/handlers/admin_dashboard_test.go`:
  - `TestAdminDashboard_ApprovePendingTransaction`
  - `TestAdminDashboard_DenyPendingTransaction`
  - `TestAdminDashboard_PendingQueueAndHistory`
- Middleware hardening checks are in:
  - `backend/internal/middleware/request_id_test.go`
  - `backend/internal/middleware/rate_limit_test.go`
- Phase 2 payment integration is covered in:
  - `backend/internal/handlers/payment_integration_test.go`
  - `backend/internal/payments/stripe_test.go`

## Useful Debug Commands

```bash
docker logs agentpay-db
lsof -i :8080
```
