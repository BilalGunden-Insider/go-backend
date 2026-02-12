package postgres

import (
	"context"
	"fmt"

	"github.com/BilalGunden-Insider/go-backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type BalanceRepository struct {
	pool *pgxpool.Pool
}

func NewBalanceRepository(pool *pgxpool.Pool) *BalanceRepository {
	return &BalanceRepository{pool: pool}
}

func (r *BalanceRepository) Create(ctx context.Context, balance *models.Balance) error {
	query := `INSERT INTO balances (user_id, amount, last_updated_at) VALUES ($1, $2, $3)`
	_, err := r.pool.Exec(ctx, query, balance.UserID, balance.Amount, balance.LastUpdatedAt)
	if err != nil {
		return fmt.Errorf("insert balance: %w", err)
	}
	return nil
}

func (r *BalanceRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Balance, error) {
	query := `SELECT user_id, amount, last_updated_at FROM balances WHERE user_id = $1`
	var b models.Balance
	err := r.pool.QueryRow(ctx, query, userID).Scan(&b.UserID, &b.Amount, &b.LastUpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan balance: %w", err)
	}
	return &b, nil
}

func (r *BalanceRepository) GetByUserIDForUpdate(ctx context.Context, dbTx pgx.Tx, userID uuid.UUID) (*models.Balance, error) {
	query := `SELECT user_id, amount, last_updated_at FROM balances WHERE user_id = $1 FOR UPDATE`
	var b models.Balance
	err := dbTx.QueryRow(ctx, query, userID).Scan(&b.UserID, &b.Amount, &b.LastUpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan balance for update: %w", err)
	}
	return &b, nil
}

func (r *BalanceRepository) UpdateAmount(ctx context.Context, dbTx pgx.Tx, userID uuid.UUID, amount decimal.Decimal) error {
	query := `UPDATE balances SET amount = $1, last_updated_at = NOW() WHERE user_id = $2`
	_, err := dbTx.Exec(ctx, query, amount, userID)
	if err != nil {
		return fmt.Errorf("update balance: %w", err)
	}
	return nil
}

func (r *BalanceRepository) GetAll(ctx context.Context) ([]*models.Balance, error) {
	rows, err := r.pool.Query(ctx, `SELECT user_id, amount, last_updated_at FROM balances`)
	if err != nil {
		return nil, fmt.Errorf("query balances: %w", err)
	}
	defer rows.Close()

	var balances []*models.Balance
	for rows.Next() {
		var b models.Balance
		if err := rows.Scan(&b.UserID, &b.Amount, &b.LastUpdatedAt); err != nil {
			return nil, fmt.Errorf("scan balance row: %w", err)
		}
		balances = append(balances, &b)
	}
	return balances, rows.Err()
}
