package postgres

import (
	"context"
	"fmt"

	"github.com/BilalGunden-Insider/go-backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionLimitRepository struct {
	pool *pgxpool.Pool
}

func NewTransactionLimitRepository(pool *pgxpool.Pool) *TransactionLimitRepository {
	return &TransactionLimitRepository{pool: pool}
}

func (r *TransactionLimitRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.TransactionLimit, error) {
	query := `SELECT id, user_id, max_per_transaction, max_daily_amount, created_at, updated_at
		FROM transaction_limits WHERE user_id = $1`

	var l models.TransactionLimit
	var uid *uuid.UUID
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&l.ID, &uid, &l.MaxPerTransaction, &l.MaxDailyAmount, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return r.GetGlobal(ctx)
		}
		return nil, fmt.Errorf("get limit by user: %w", err)
	}
	l.UserID = uid
	return &l, nil
}

func (r *TransactionLimitRepository) GetGlobal(ctx context.Context) (*models.TransactionLimit, error) {
	query := `SELECT id, user_id, max_per_transaction, max_daily_amount, created_at, updated_at
		FROM transaction_limits WHERE user_id IS NULL LIMIT 1`

	var l models.TransactionLimit
	var uid *uuid.UUID
	err := r.pool.QueryRow(ctx, query).Scan(
		&l.ID, &uid, &l.MaxPerTransaction, &l.MaxDailyAmount, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get global limit: %w", err)
	}
	l.UserID = uid
	return &l, nil
}

func (r *TransactionLimitRepository) Upsert(ctx context.Context, limit *models.TransactionLimit) error {
	query := `
		INSERT INTO transaction_limits (id, user_id, max_per_transaction, max_daily_amount)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) WHERE user_id IS NOT NULL
		DO UPDATE SET max_per_transaction = $3, max_daily_amount = $4, updated_at = NOW()`
	_, err := r.pool.Exec(ctx, query, limit.ID, limit.UserID, limit.MaxPerTransaction, limit.MaxDailyAmount)
	if err != nil {
		return fmt.Errorf("upsert limit: %w", err)
	}
	return nil
}
