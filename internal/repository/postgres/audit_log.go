package postgres

import (
	"context"
	"fmt"

	"github.com/BilalGunden-Insider/go-backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditLogRepository struct {
	pool *pgxpool.Pool
}

func NewAuditLogRepository(pool *pgxpool.Pool) *AuditLogRepository {
	return &AuditLogRepository{pool: pool}
}

func (r *AuditLogRepository) Create(ctx context.Context, log *models.AuditLog) error {
	query := `INSERT INTO audit_logs (id, entity_type, entity_id, action, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.pool.Exec(ctx, query,
		log.ID, log.EntityType, log.EntityID, log.Action, log.Details, log.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

func (r *AuditLogRepository) ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]*models.AuditLog, error) {
	query := `SELECT id, entity_type, entity_id, action, details, created_at
		FROM audit_logs
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.AuditLog
	for rows.Next() {
		var l models.AuditLog
		if err := rows.Scan(&l.ID, &l.EntityType, &l.EntityID, &l.Action, &l.Details, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan audit log row: %w", err)
		}
		logs = append(logs, &l)
	}
	return logs, rows.Err()
}
