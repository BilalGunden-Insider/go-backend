package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/BilalGunden-Insider/go-backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type TransactionRepository struct {
	pool *pgxpool.Pool
}

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{pool: pool}
}

func nullableUUID(id uuid.UUID) interface{} {
	if id == uuid.Nil {
		return nil
	}
	return id
}

func scanNullableUUID(src pgtype.UUID) uuid.UUID {
	if !src.Valid {
		return uuid.Nil
	}
	return uuid.UUID(src.Bytes)
}

func (r *TransactionRepository) Create(ctx context.Context, tx *models.Transaction) error {
	query := `
		INSERT INTO transactions (id, from_user_id, to_user_id, amount, type, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.pool.Exec(ctx, query,
		tx.ID, nullableUUID(tx.FromUserID), nullableUUID(tx.ToUserID), tx.Amount, tx.Type, tx.Status, tx.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}
	return nil
}

func (r *TransactionRepository) CreateTx(ctx context.Context, dbTx pgx.Tx, tx *models.Transaction) error {
	query := `
		INSERT INTO transactions (id, from_user_id, to_user_id, amount, type, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := dbTx.Exec(ctx, query,
		tx.ID, nullableUUID(tx.FromUserID), nullableUUID(tx.ToUserID), tx.Amount, tx.Type, tx.Status, tx.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert transaction in tx: %w", err)
	}
	return nil
}

func (r *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	query := `SELECT id, from_user_id, to_user_id, amount, type, status, created_at
		FROM transactions WHERE id = $1`
	var t models.Transaction
	var fromUID, toUID pgtype.UUID
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &fromUID, &toUID, &t.Amount, &t.Type, &t.Status, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan transaction: %w", err)
	}
	t.FromUserID = scanNullableUUID(fromUID)
	t.ToUserID = scanNullableUUID(toUID)
	return &t, nil
}

func (r *TransactionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE transactions SET status = $1 WHERE id = $2`, status, id)
	if err != nil {
		return fmt.Errorf("update transaction status: %w", err)
	}
	return nil
}

func (r *TransactionRepository) UpdateStatusTx(ctx context.Context, dbTx pgx.Tx, id uuid.UUID, status string) error {
	_, err := dbTx.Exec(ctx,
		`UPDATE transactions SET status = $1 WHERE id = $2`, status, id)
	if err != nil {
		return fmt.Errorf("update transaction status in tx: %w", err)
	}
	return nil
}

func (r *TransactionRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Transaction, error) {
	query := `SELECT id, from_user_id, to_user_id, amount, type, status, created_at
		FROM transactions
		WHERE from_user_id = $1 OR to_user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query transactions: %w", err)
	}
	defer rows.Close()

	var txns []*models.Transaction
	for rows.Next() {
		var t models.Transaction
		var fromUID, toUID pgtype.UUID
		if err := rows.Scan(&t.ID, &fromUID, &toUID, &t.Amount, &t.Type, &t.Status, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan transaction row: %w", err)
		}
		t.FromUserID = scanNullableUUID(fromUID)
		t.ToUserID = scanNullableUUID(toUID)
		txns = append(txns, &t)
	}
	return txns, rows.Err()
}

func (r *TransactionRepository) GetDailyTotal(ctx context.Context, userID uuid.UUID, date time.Time) (decimal.Decimal, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.AddDate(0, 0, 1)

	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE from_user_id = $1 AND status = 'completed'
		  AND created_at >= $2 AND created_at < $3`

	var total decimal.Decimal
	err := r.pool.QueryRow(ctx, query, userID, startOfDay, endOfDay).Scan(&total)
	if err != nil {
		return decimal.Zero, fmt.Errorf("get daily total: %w", err)
	}
	return total, nil
}

func (r *TransactionRepository) CalculateBalanceAt(ctx context.Context, userID uuid.UUID, at time.Time) (decimal.Decimal, error) {
	query := `
		SELECT COALESCE(SUM(CASE WHEN to_user_id = $1 THEN amount ELSE -amount END), 0)
		FROM transactions
		WHERE (to_user_id = $1 OR from_user_id = $1)
		  AND status = 'completed'
		  AND created_at <= $2`
	var result decimal.Decimal
	err := r.pool.QueryRow(ctx, query, userID, at).Scan(&result)
	if err != nil {
		return decimal.Zero, fmt.Errorf("calculate balance at: %w", err)
	}
	return result, nil
}
