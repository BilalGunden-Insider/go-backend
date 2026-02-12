package repository

import (
	"context"

	"github.com/bilal/backend_path/internal/models"
	"github.com/google/uuid"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log *models.AuditLog) error
	ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]*models.AuditLog, error)
}
