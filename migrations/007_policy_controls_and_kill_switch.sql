-- Migration 007: Structured policy controls, hard spend constraints, and org kill switch support.

ALTER TABLE users
ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'users_status_check'
    ) THEN
        ALTER TABLE users
        ADD CONSTRAINT users_status_check CHECK (status IN ('active', 'frozen'));
    END IF;
END $$;

ALTER TABLE policies
ADD COLUMN IF NOT EXISTS per_transaction_limit_cents BIGINT NOT NULL DEFAULT 50000;

ALTER TABLE policies
ADD COLUMN IF NOT EXISTS allowed_mccs TEXT[] NOT NULL DEFAULT '{}'::TEXT[];

ALTER TABLE policies
ADD COLUMN IF NOT EXISTS allowed_weekdays_utc INTEGER[] NOT NULL DEFAULT ARRAY[0,1,2,3,4,5,6]::INTEGER[];

ALTER TABLE policies
ADD COLUMN IF NOT EXISTS allowed_hours_utc INTEGER[] NOT NULL DEFAULT ARRAY[
    0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23
]::INTEGER[];

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'policies_daily_limit_hard_max'
    ) THEN
        ALTER TABLE policies
        ADD CONSTRAINT policies_daily_limit_hard_max
        CHECK (daily_limit_cents > 0 AND daily_limit_cents <= 100000000);
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'policies_per_transaction_limit_hard_max'
    ) THEN
        ALTER TABLE policies
        ADD CONSTRAINT policies_per_transaction_limit_hard_max
        CHECK (per_transaction_limit_cents > 0 AND per_transaction_limit_cents <= 1000000);
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'policies_approval_threshold_bounds'
    ) THEN
        ALTER TABLE policies
        ADD CONSTRAINT policies_approval_threshold_bounds
        CHECK (
            require_approval_above_cents >= 0
            AND require_approval_above_cents <= per_transaction_limit_cents
        );
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'transactions_amount_hard_max'
    ) THEN
        ALTER TABLE transactions
        ADD CONSTRAINT transactions_amount_hard_max
        CHECK (amount_cents > 0 AND amount_cents <= 1000000);
    END IF;
END $$;
