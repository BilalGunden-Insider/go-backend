package repository

import (
	"context"
	"time"

	"github.com/BilalGunden-Insider/go-backend/internal/models"
	"github.com/google/uuid"
)

type ScheduledTransactionRepository interface {
	Create(ctx context.Context, st *models.ScheduledTransaction) error
	GetDue(ctx context.Context, now time.Time, limit int) ([]*models.ScheduledTransaction, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, errMsg string) error
	SetExecuted(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.ScheduledTransaction, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.ScheduledTransaction, error)
	Cancel(ctx context.Context, id uuid.UUID) error
}
