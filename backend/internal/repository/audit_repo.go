package repository

import (
	"context"
	"database/sql"
	"fmt"

	"agentpay/internal/models"
)

// AuditRepository persists review audit events.
type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) CreateApprovalLog(ctx context.Context, tx *sql.Tx, req *models.CreateApprovalAuditLogRequest) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO approval_audit_logs (
			transaction_id,
			admin_id,
			action,
			previous_status,
			new_status,
			reason,
			request_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''))
	`, req.TransactionID, req.AdminID, req.Action, req.PreviousStatus, req.NewStatus, req.Reason, req.RequestID)
	if err != nil {
		return fmt.Errorf("failed to create approval audit log: %w", err)
	}
	return nil
}

func (r *AuditRepository) ListApprovalLogs(ctx context.Context, limit int) ([]models.ApprovalAuditLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			l.id,
			l.transaction_id,
			l.admin_id,
			a.email,
			l.action,
			l.previous_status,
			l.new_status,
			l.reason,
			COALESCE(l.request_id, ''),
			l.created_at
		FROM approval_audit_logs l
		JOIN admins a ON a.id = l.admin_id
		ORDER BY l.created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list approval audit logs: %w", err)
	}
	defer rows.Close()

	logs := make([]models.ApprovalAuditLog, 0, limit)
	for rows.Next() {
		var item models.ApprovalAuditLog
		if err := rows.Scan(
			&item.ID,
			&item.TransactionID,
			&item.AdminID,
			&item.AdminEmail,
			&item.Action,
			&item.PreviousStatus,
			&item.NewStatus,
			&item.Reason,
			&item.RequestID,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan approval audit log: %w", err)
		}
		logs = append(logs, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterating approval audit logs: %w", err)
	}

	return logs, nil
}
