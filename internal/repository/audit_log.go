package repository

import (
	"context"

	"github.com/BilalGunden-Insider/go-backend/internal/models"
	"github.com/google/uuid"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log *models.AuditLog) error
	ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]*models.AuditLog, error)
}
