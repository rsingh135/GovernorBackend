-- Migration 005: Add approval audit logging for admin review actions.

CREATE TABLE IF NOT EXISTS approval_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    admin_id UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    action VARCHAR(16) NOT NULL CHECK (action IN ('approve', 'deny')),
    previous_status VARCHAR(32) NOT NULL,
    new_status VARCHAR(32) NOT NULL,
    reason TEXT NOT NULL,
    request_id VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_approval_audit_logs_transaction_id ON approval_audit_logs(transaction_id);
CREATE INDEX IF NOT EXISTS idx_approval_audit_logs_admin_id ON approval_audit_logs(admin_id);
CREATE INDEX IF NOT EXISTS idx_approval_audit_logs_created_at ON approval_audit_logs(created_at DESC);
