package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/BilalGunden-Insider/go-backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScheduledTransactionRepository struct {
	pool *pgxpool.Pool
}

func NewScheduledTransactionRepository(pool *pgxpool.Pool) *ScheduledTransactionRepository {
	return &ScheduledTransactionRepository{pool: pool}
}

func (r *ScheduledTransactionRepository) Create(ctx context.Context, st *models.ScheduledTransaction) error {
	query := `
		INSERT INTO scheduled_transactions (id, from_user_id, to_user_id, amount, type, status, scheduled_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.pool.Exec(ctx, query,
		st.ID, nullableUUID(st.FromUserID), nullableUUID(st.ToUserID),
		st.Amount, st.Type, st.Status, st.ScheduledAt, st.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert scheduled transaction: %w", err)
	}
	return nil
}

func (r *ScheduledTransactionRepository) GetDue(ctx context.Context, now time.Time, limit int) ([]*models.ScheduledTransaction, error) {
	query := `
		SELECT id, from_user_id, to_user_id, amount, type, status, scheduled_at, created_at
		FROM scheduled_transactions
		WHERE status = 'pending' AND scheduled_at <= $1
		ORDER BY scheduled_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED`

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, query, now, limit)
	if err != nil {
		return nil, fmt.Errorf("query due scheduled transactions: %w", err)
	}
	defer rows.Close()

	var result []*models.ScheduledTransaction
	for rows.Next() {
		var st models.ScheduledTransaction
		var fromUID, toUID pgtype.UUID
		if err := rows.Scan(&st.ID, &fromUID, &toUID, &st.Amount, &st.Type, &st.Status, &st.ScheduledAt, &st.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan scheduled transaction: %w", err)
		}
		st.FromUserID = scanNullableUUID(fromUID)
		st.ToUserID = scanNullableUUID(toUID)
		result = append(result, &st)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, rows.Err()
}

func (r *ScheduledTransactionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, errMsg string) error {
	query := `UPDATE scheduled_transactions SET status = $1, error_message = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, status, errMsg, id)
	if err != nil {
		return fmt.Errorf("update scheduled transaction status: %w", err)
	}
	return nil
}

func (r *ScheduledTransactionRepository) SetExecuted(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	query := `UPDATE scheduled_transactions SET status = 'executed', executed_at = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("set scheduled transaction executed: %w", err)
	}
	return nil
}

func (r *ScheduledTransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ScheduledTransaction, error) {
	query := `SELECT id, from_user_id, to_user_id, amount, type, status, scheduled_at, executed_at, error_message, created_at
		FROM scheduled_transactions WHERE id = $1`

	var st models.ScheduledTransaction
	var fromUID, toUID pgtype.UUID
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&st.ID, &fromUID, &toUID, &st.Amount, &st.Type, &st.Status,
		&st.ScheduledAt, &st.ExecutedAt, &st.ErrorMessage, &st.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get scheduled transaction: %w", err)
	}
	st.FromUserID = scanNullableUUID(fromUID)
	st.ToUserID = scanNullableUUID(toUID)
	return &st, nil
}

func (r *ScheduledTransactionRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.ScheduledTransaction, error) {
	query := `SELECT id, from_user_id, to_user_id, amount, type, status, scheduled_at, executed_at, error_message, created_at
		FROM scheduled_transactions
		WHERE from_user_id = $1 OR to_user_id = $1
		ORDER BY scheduled_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list scheduled transactions: %w", err)
	}
	defer rows.Close()

	var result []*models.ScheduledTransaction
	for rows.Next() {
		var st models.ScheduledTransaction
		var fromUID, toUID pgtype.UUID
		if err := rows.Scan(&st.ID, &fromUID, &toUID, &st.Amount, &st.Type, &st.Status,
			&st.ScheduledAt, &st.ExecutedAt, &st.ErrorMessage, &st.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan scheduled transaction: %w", err)
		}
		st.FromUserID = scanNullableUUID(fromUID)
		st.ToUserID = scanNullableUUID(toUID)
		result = append(result, &st)
	}
	return result, rows.Err()
}

func (r *ScheduledTransactionRepository) Cancel(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE scheduled_transactions SET status = 'cancelled' WHERE id = $1 AND status = 'pending'`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("cancel scheduled transaction: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("scheduled transaction not found or not pending")
	}
	return nil
}
