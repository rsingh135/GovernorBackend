-- AgentPay MVP Database Schema (Phase 1)
-- CRITICAL: All monetary values are stored as integers representing cents.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Agents are the authenticated principals for /spend.
-- api_key is NEVER stored in plaintext; we store sha256(api_key) as api_key_hash.
CREATE TABLE agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'frozen')),
    api_key_hash BYTEA UNIQUE NOT NULL,
    api_key_prefix VARCHAR(16) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- One policy per agent (upsert behavior in API).
CREATE TABLE policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID UNIQUE NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    currency CHAR(3) NOT NULL DEFAULT 'usd',
    daily_limit_cents BIGINT NOT NULL,
    allowed_vendors TEXT[] NOT NULL DEFAULT '{}'::TEXT[],
    require_approval_above_cents BIGINT NOT NULL DEFAULT 0,
    raw_policy JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Transactions are the idempotent ledger of /spend requests.
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    request_id UUID UNIQUE NOT NULL,
    amount_cents BIGINT NOT NULL,
    currency CHAR(3) NOT NULL,
    vendor TEXT NOT NULL,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(32) NOT NULL CHECK (status IN ('APPROVED', 'DENIED', 'PENDING_APPROVAL')),
    reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes for /spend hot path.
CREATE INDEX idx_policies_agent_id ON policies(agent_id);
CREATE INDEX idx_transactions_agent_id_created_at ON transactions(agent_id, created_at);
CREATE INDEX idx_transactions_request_id ON transactions(request_id);

-- Seed Data (dev only)
-- apiKey for the seeded agent: sk_test_agent_123
INSERT INTO agents (id, name, status, api_key_hash, api_key_prefix) VALUES
    ('22222222-2222-2222-2222-222222222222', 'TigerBot', 'active', digest('sk_test_agent_123', 'sha256'), 'sk_test_');

INSERT INTO policies (agent_id, currency, daily_limit_cents, allowed_vendors, require_approval_above_cents, raw_policy) VALUES
    (
        '22222222-2222-2222-2222-222222222222',
        'usd',
        10000,
        ARRAY['openai.com']::TEXT[],
        1500,
        jsonb_build_object(
            'currency', 'usd',
            'daily_limit_cents', 10000,
            'allowed_vendors', jsonb_build_array('openai.com'),
            'require_approval_above_cents', 1500
        )
    );
