# Phase 2: Production-Ready Features & External Integrations

## Current Status: Phase 1 Complete ✅

**Phase 1 (Complete - Feb 2026)**
- ✅ Modular 4-layer Go architecture (Handlers → Services → Repositories → Database)
- ✅ User accounts with balance tracking (`users` table)
- ✅ Agent provisioning with SHA256-hashed API keys
- ✅ Policy management (daily limits, vendor allowlists, approval thresholds)
- ✅ Idempotent spending engine with row-level locking (`SELECT FOR UPDATE`)
- ✅ Integer-only currency (no floats)
- ✅ Comprehensive test suite (5 tests, 100% pass rate)
- ✅ API Endpoints: `POST /users`, `POST /agents`, `POST /policies`, `POST /spend`
- ✅ Documentation: `PHASE_1_COMPLETE.md`, `QUICK_START.md`, `TESTING_GUIDE.md`

**Lines of Code**: ~2,024 lines | **Build**: Clean compilation | **Tests**: 5/5 passing

---

## Phase 2 Roadmap (Planned)

Transform Governor from a spending engine into a complete financial middleware platform.

### Why Phase 2?
Phase 1 lacks:
- **Visibility**: No GET endpoints to query data
- **Workflow Completion**: Pending transactions stay pending forever
- **Observability**: No structured logging for debugging/auditing
- **Protection**: No rate limiting against abuse
- **Integrations**: No webhooks for external systems
- **Error Correction**: No way to reverse erroneous transactions

---

## Implementation Phases

### 🟦 Phase 2A: GET Endpoints (Foundation)
**Complexity**: Medium | **Effort**: 3-4 days | **Dependencies**: None

Add read API surface to query all entities.

**Features**:
- [ ] `GET /users/:id` - Retrieve user by ID
- [ ] `GET /agents?user_id=&status=&limit=100` - List agents with filters
- [ ] `GET /transactions?agent_id=&status=&from_date=&limit=100` - List transactions with pagination
- [ ] `GET /policies/:agent_id` - Get policy for agent

**Database Changes**:
- [ ] Add indexes for efficient listing queries (`idx_transactions_status`, `idx_transactions_created_at`, etc.)

**New Files**:
- [ ] `internal/models/pagination.go` - Pagination helpers
- [ ] `internal/models/filters.go` - Filter structs
- [ ] `internal/httputil/query.go` - Query param parsing utilities
- [ ] `internal/handlers/transaction_handler.go` - Transaction listing handler
- [ ] `internal/services/transaction_service.go` - Transaction listing service

**Security**:
- Enforce auth scoping (agents only see own transactions)
- Max limit enforcement (1000) to prevent DoS

---

### 🟩 Phase 2B: Structured Logging (Observability)
**Complexity**: Low-Medium | **Effort**: 2-3 days | **Dependencies**: None

Add structured JSON logging with `zerolog` for operational visibility.

**Features**:
- [ ] Add `zerolog` dependency
- [ ] Logger middleware with request ID tracking
- [ ] Contextual logging (request_id, agent_id, user_id in all logs)
- [ ] Log all critical business events (user created, spend approved/denied, etc.)
- [ ] Environment-based log levels (DEBUG, INFO, WARN, ERROR)

**New Files**:
- [ ] `internal/logger/logger.go` - Logger initialization
- [ ] `internal/middleware/logger.go` - Logger middleware

**Log Events**:
- Request lifecycle (received, completed with status/duration)
- Business events (user created, spend processed, approval granted)
- Error events (database errors, validation failures)

---

### 🟨 Phase 2C: Approval Workflow (Core Business Logic)
**Complexity**: High | **Effort**: 4-5 days | **Dependencies**: Phase 2A

Complete the transaction lifecycle with approval/denial capabilities.

**Features**:
- [ ] `POST /transactions/:id/approve` - Approve pending transaction
- [ ] `POST /transactions/:id/deny` - Deny pending transaction
- [ ] Balance deduction at approval time (not at spend time for pending txns)
- [ ] Audit trail (`approval_actions` table)
- [ ] Authorization (only users with `can_approve=true` or agent's owner)

**Database Changes**:
- [ ] Add `can_approve` column to `users` table
- [ ] Add `approved_at`, `approved_by_user_id` columns to `transactions` table
- [ ] Create `approval_actions` table for audit trail

**New Files**:
- [ ] `internal/models/approval.go` - Approval models
- [ ] `internal/handlers/approval_handler.go` - Approval HTTP handlers
- [ ] `internal/services/approval_service.go` - Approval business logic
- [ ] `internal/repository/approval_repo.go` - Approval data access

**Business Logic**:
1. Verify user has approval permission
2. Lock transaction row (`FOR UPDATE`)
3. Verify status is `PENDING_APPROVAL`
4. Lock user row and re-check balance
5. Update transaction status to `APPROVED`/`DENIED`
6. Deduct balance (if approving)
7. Record action in audit trail

---

### 🟧 Phase 2D: Rate Limiting (API Protection)
**Complexity**: Medium | **Effort**: 2-3 days | **Dependencies**: None

Protect API from abuse with token bucket rate limiting.

**Features**:
- [ ] Per-API-key rate limiting (100 req/sec for authenticated)
- [ ] Per-IP rate limiting (10 req/sec for public endpoints)
- [ ] Configurable limits per endpoint type
- [ ] Standard rate limit headers (`X-RateLimit-Limit`, `X-RateLimit-Remaining`)
- [ ] 429 Too Many Requests response when exceeded

**Dependencies**:
- [ ] Add `golang.org/x/time` dependency

**New Files**:
- [ ] `internal/ratelimit/limiter.go` - Token bucket limiter
- [ ] `internal/middleware/ratelimit.go` - Rate limit middleware

**Rate Limits**:
- Public endpoints: 10 req/sec per IP
- Authenticated endpoints: 100 req/sec per API key
- Approval endpoints: 10 req/sec per API key

---

### 🟪 Phase 2E: Webhook Notifications (External Integrations)
**Complexity**: High | **Effort**: 4-5 days | **Dependencies**: Phase 2C

Enable external systems to react to approval/denial events.

**Features**:
- [ ] `POST /webhooks` - Register webhook endpoint
- [ ] `GET /webhooks` - List registered webhooks
- [ ] `DELETE /webhooks/:id` - Delete webhook
- [ ] `GET /webhooks/:id/deliveries` - View delivery history
- [ ] Webhook delivery with HMAC signature verification
- [ ] Retry logic with exponential backoff (1min, 5min, 15min, 1hr, 6hr, 24hr)
- [ ] Background worker for retrying failed deliveries

**Database Changes**:
- [ ] Create `webhooks` table
- [ ] Create `webhook_deliveries` table

**New Files**:
- [ ] `internal/models/webhook.go` - Webhook models
- [ ] `internal/webhook/signer.go` - HMAC signature utilities
- [ ] `internal/handlers/webhook_handler.go` - Webhook HTTP handlers
- [ ] `internal/services/webhook_service.go` - Webhook delivery & retry logic
- [ ] `internal/repository/webhook_repo.go` - Webhook data access
- [ ] `internal/repository/webhook_delivery_repo.go` - Delivery tracking

**Event Types**:
- `transaction.approved` - Triggered when transaction is approved
- `transaction.denied` - Triggered when transaction is denied

**Security**:
- HMAC-SHA256 signatures prevent webhook spoofing
- Webhooks user-scoped (only receive events for their agents)
- HTTPS-only webhook URLs (production)

---

### 🟥 Phase 2F: Transaction Reversals (Error Correction)
**Complexity**: High | **Effort**: 3-4 days | **Dependencies**: Phase 2A, Phase 2C

Enable reversing erroneous approved transactions with balance restoration.

**Features**:
- [ ] `POST /transactions/:id/reverse` - Reverse approved transaction
- [ ] `GET /reversals` - List reversals with filters
- [ ] Balance restoration (add back to user balance)
- [ ] Reversal reasons (customer_request, fraud, error, duplicate, other)
- [ ] Idempotency via `idempotency_key` in request body
- [ ] Audit trail for all reversals

**Database Changes**:
- [ ] Create `reversals` table
- [ ] Add `reversed_at`, `reversal_id` columns to `transactions` table

**New Files**:
- [ ] `internal/models/reversal.go` - Reversal models
- [ ] `internal/handlers/reversal_handler.go` - Reversal HTTP handlers
- [ ] `internal/services/reversal_service.go` - Reversal business logic
- [ ] `internal/repository/reversal_repo.go` - Reversal data access

**Business Logic**:
1. Check idempotency (if reversal exists with same key, return it)
2. Lock transaction row (`FOR UPDATE`)
3. Verify status is `APPROVED` (not PENDING or DENIED)
4. Verify not already reversed
5. Lock user row
6. Restore balance: `balance_cents = balance_cents + amount`
7. Create reversal record
8. Update transaction with `reversed_at` and `reversal_id`

---

## Implementation Timeline

**Recommended: Parallel Track Approach**

| Week | Phases | Rationale |
|------|--------|-----------|
| **Week 1** | 2A (GET endpoints) + 2B (Logging) | Independent, can run in parallel |
| **Week 2** | 2C (Approvals) + 2D (Rate limiting) | 2C depends on 2A; 2D is independent |
| **Week 3** | 2E (Webhooks) + 2F (Reversals) | 2E depends on 2C; 2F depends on 2A & 2C |
| **Week 4** | Testing & refinement | Integration testing, documentation |

**Estimated Total**: 4-5 weeks

---

## Success Criteria

- [ ] **2A**: All GET endpoints return paginated data with auth scoping
- [ ] **2B**: All requests logged with request_id, business events captured in structured logs
- [ ] **2C**: Can approve/deny pending transactions, balance deducted correctly, audit trail complete
- [ ] **2D**: Rate limits enforced, 429 responses when exceeded, headers present
- [ ] **2E**: Webhooks triggered on approval, retries work, HMAC signatures valid
- [ ] **2F**: Can reverse approved transactions, balance restored, idempotent

---

## Technical Notes

- All currency operations remain `int64` (cents) - no floats
- All queries remain parameterized - SQL injection safe
- Row-level locking preserved for critical operations
- 100% backward compatibility with Phase 1 APIs
- Follow existing patterns: same folder structure, same DI approach, same testing strategy

---

## Related Documentation

- **Detailed Implementation Plan**: `/Users/ranveersingh/.claude/plans/crispy-nibbling-peacock.md`
- **Phase 1 Documentation**: `PHASE_1_COMPLETE.md`, `QUICK_START.md`
- **Current Test Suite**: `backend/internal/handlers/api_test.go`

---

**Labels**: `enhancement`, `phase-2`, `production-ready`
**Assignees**: @rsingh135
**Milestone**: Phase 2 - Production Features
