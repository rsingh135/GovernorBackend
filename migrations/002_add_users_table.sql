-- Migration 002: Add Users Table and Refactor Agents
-- This migration adds user accounts with balance tracking and links agents to users.

-- Create users table (account holders with balances)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    balance_cents BIGINT NOT NULL DEFAULT 0 CHECK (balance_cents >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Add user_id column to agents table (if not exists)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name='agents' AND column_name='user_id'
    ) THEN
        ALTER TABLE agents ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Create index for user lookups
CREATE INDEX IF NOT EXISTS idx_agents_user_id ON agents(user_id);

-- Create index for user balance operations (will be used with FOR UPDATE)
CREATE INDEX IF NOT EXISTS idx_users_id_balance ON users(id, balance_cents);

-- Seed test user data (for development/testing)
INSERT INTO users (id, name, balance_cents) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Ranveer', 5000)
ON CONFLICT (id) DO NOTHING;

-- Update existing BhangraBot agent to link to Ranveer user
UPDATE agents
SET user_id = '11111111-1111-1111-1111-111111111111'
WHERE id = '22222222-2222-2222-2222-222222222222';
