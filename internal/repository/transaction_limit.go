package repository

import (
	"context"

	"github.com/BilalGunden-Insider/go-backend/internal/models"
	"github.com/google/uuid"
)

type TransactionLimitRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.TransactionLimit, error)
	GetGlobal(ctx context.Context) (*models.TransactionLimit, error)
	Upsert(ctx context.Context, limit *models.TransactionLimit) error
}
