-- Migration 003: Add indexes for efficient listing queries
-- Phase 2A: GET Endpoints

-- Transaction listing indexes
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_agent_status_created ON transactions(agent_id, status, created_at DESC);

-- Agent listing indexes
CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
CREATE INDEX IF NOT EXISTS idx_agents_user_id ON agents(user_id);
