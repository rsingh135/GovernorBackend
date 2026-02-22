-- Migration 003: Add status column to users (aligns with repositories and services)

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'frozen'));

-- Backfill any preexisting rows where column was added without default
UPDATE users SET status = 'active' WHERE status IS NULL;
