-- Migration 004: Add per-transaction limits and allowlist fields to policies

ALTER TABLE policies
    ADD COLUMN IF NOT EXISTS per_transaction_limit_cents BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS allowed_mccs TEXT[] NOT NULL DEFAULT '{}'::TEXT[],
    ADD COLUMN IF NOT EXISTS allowed_weekdays_utc INT[] NOT NULL DEFAULT '{}'::INT[],
    ADD COLUMN IF NOT EXISTS allowed_hours_utc INT[] NOT NULL DEFAULT '{}'::INT[],
    ADD COLUMN IF NOT EXISTS purchase_guideline TEXT NOT NULL DEFAULT '';

-- Backfill per_transaction_limit_cents to current daily_limit_cents where zero/NULL
UPDATE policies
SET per_transaction_limit_cents = daily_limit_cents
WHERE per_transaction_limit_cents IS NULL OR per_transaction_limit_cents = 0;

-- Ensure require_approval_above_cents does not exceed per_transaction_limit_cents
UPDATE policies
SET require_approval_above_cents = per_transaction_limit_cents
WHERE require_approval_above_cents > per_transaction_limit_cents;
