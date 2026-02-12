package repository

import (
	"context"

	"github.com/bilal/backend_path/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type BalanceRepository interface {
	Create(ctx context.Context, balance *models.Balance) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Balance, error)
	GetByUserIDForUpdate(ctx context.Context, dbTx pgx.Tx, userID uuid.UUID) (*models.Balance, error)
	UpdateAmount(ctx context.Context, dbTx pgx.Tx, userID uuid.UUID, amount decimal.Decimal) error
	GetAll(ctx context.Context) ([]*models.Balance, error)
}
