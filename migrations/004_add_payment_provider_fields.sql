-- Migration 004: Add payment-provider tracking to transactions and webhook idempotency table.

ALTER TABLE transactions ADD COLUMN IF NOT EXISTS provider VARCHAR(32);
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS provider_session_id VARCHAR(255);
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS provider_payment_intent_id VARCHAR(255);
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS provider_status VARCHAR(64) NOT NULL DEFAULT 'not_applicable';
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS provider_checkout_url TEXT;

CREATE INDEX IF NOT EXISTS idx_transactions_provider_session_id ON transactions(provider_session_id);
CREATE INDEX IF NOT EXISTS idx_transactions_provider_payment_intent_id ON transactions(provider_payment_intent_id);

CREATE TABLE IF NOT EXISTS payment_webhook_events (
    event_id VARCHAR(255) PRIMARY KEY,
    provider VARCHAR(32) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
