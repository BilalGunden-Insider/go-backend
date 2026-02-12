package repository

import (
	"context"

	"github.com/BilalGunden-Insider/go-backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *models.Transaction) error
	CreateTx(ctx context.Context, dbTx pgx.Tx, tx *models.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateStatusTx(ctx context.Context, dbTx pgx.Tx, id uuid.UUID, status string) error
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Transaction, error)
}
