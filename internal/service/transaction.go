package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/bilal/backend_path/internal/models"
	"github.com/bilal/backend_path/internal/repository"
	"github.com/bilal/backend_path/internal/worker"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type TransactionService struct {
	txns     repository.TransactionRepository
	audit    repository.AuditLogRepository
	balances *BalanceService
	pool     *pgxpool.Pool
	wp       *worker.Pool
	log      *slog.Logger
}

func NewTransactionService(
	txns repository.TransactionRepository,
	audit repository.AuditLogRepository,
	balances *BalanceService,
	pool *pgxpool.Pool,
	log *slog.Logger,
) *TransactionService {
	return &TransactionService{
		txns:     txns,
		audit:    audit,
		balances: balances,
		pool:     pool,
		log:      log,
	}
}

func (s *TransactionService) SetWorkerPool(wp *worker.Pool) {
	s.wp = wp
}

func (s *TransactionService) Credit(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (*models.Transaction, error) {
	tx := &models.Transaction{
		ID:         uuid.New(),
		FromUserID: uuid.Nil,
		ToUserID:   userID,
		Amount:     amount,
		Type:       models.TxTypeCredit,
		Status:     models.TxStatusPending,
		CreatedAt:  time.Now(),
	}
	if err := tx.Validate(); err != nil {
		return nil, err
	}

	dbTx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer dbTx.Rollback(ctx)

	if err := s.txns.CreateTx(ctx, dbTx, tx); err != nil {
		return nil, err
	}

	bal, err := s.balances.GetBalanceForUpdate(ctx, dbTx, userID)
	if err != nil {
		return nil, fmt.Errorf("lock balance: %w", err)
	}

	newAmount := bal.Amount.Add(amount)
	if err := s.balances.UpdateBalanceInTx(ctx, dbTx, userID, newAmount); err != nil {
		return nil, err
	}

	tx.Status = models.TxStatusCompleted
	if err := s.txns.UpdateStatusTx(ctx, dbTx, tx.ID, models.TxStatusCompleted); err != nil {
		return nil, err
	}

	if err := dbTx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	s.logAudit(ctx, tx)
	s.log.Info("credit completed", slog.String("user_id", userID.String()), slog.String("amount", amount.String()))
	return tx, nil
}

func (s *TransactionService) Debit(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (*models.Transaction, error) {
	tx := &models.Transaction{
		ID:         uuid.New(),
		FromUserID: userID,
		ToUserID:   uuid.Nil,
		Amount:     amount,
		Type:       models.TxTypeDebit,
		Status:     models.TxStatusPending,
		CreatedAt:  time.Now(),
	}
	if err := tx.Validate(); err != nil {
		return nil, err
	}

	dbTx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer dbTx.Rollback(ctx)

	bal, err := s.balances.GetBalanceForUpdate(ctx, dbTx, userID)
	if err != nil {
		return nil, fmt.Errorf("lock balance: %w", err)
	}
	if !bal.HasSufficientFunds(amount) {
		return nil, fmt.Errorf("insufficient funds: have %s, need %s", bal.Amount, amount)
	}

	if err := s.txns.CreateTx(ctx, dbTx, tx); err != nil {
		return nil, err
	}

	newAmount := bal.Amount.Sub(amount)
	if err := s.balances.UpdateBalanceInTx(ctx, dbTx, userID, newAmount); err != nil {
		return nil, err
	}

	tx.Status = models.TxStatusCompleted
	if err := s.txns.UpdateStatusTx(ctx, dbTx, tx.ID, models.TxStatusCompleted); err != nil {
		return nil, err
	}

	if err := dbTx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	s.logAudit(ctx, tx)
	s.log.Info("debit completed", slog.String("user_id", userID.String()), slog.String("amount", amount.String()))
	return tx, nil
}

func (s *TransactionService) Transfer(ctx context.Context, fromUserID, toUserID uuid.UUID, amount decimal.Decimal) (*models.Transaction, error) {
	tx := &models.Transaction{
		ID:         uuid.New(),
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Amount:     amount,
		Type:       models.TxTypeTransfer,
		Status:     models.TxStatusPending,
		CreatedAt:  time.Now(),
	}
	if err := tx.Validate(); err != nil {
		return nil, err
	}

	dbTx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer dbTx.Rollback(ctx)

	if err := s.txns.CreateTx(ctx, dbTx, tx); err != nil {
		return nil, err
	}

	senderBal, err := s.balances.GetBalanceForUpdate(ctx, dbTx, fromUserID)
	if err != nil {
		return nil, fmt.Errorf("lock sender balance: %w", err)
	}

	receiverBal, err := s.balances.GetBalanceForUpdate(ctx, dbTx, toUserID)
	if err != nil {
		return nil, fmt.Errorf("lock receiver balance: %w", err)
	}

	if !senderBal.HasSufficientFunds(amount) {
		return nil, fmt.Errorf("insufficient funds: have %s, need %s", senderBal.Amount, amount)
	}

	if err := s.balances.UpdateBalanceInTx(ctx, dbTx, fromUserID, senderBal.Amount.Sub(amount)); err != nil {
		return nil, err
	}
	if err := s.balances.UpdateBalanceInTx(ctx, dbTx, toUserID, receiverBal.Amount.Add(amount)); err != nil {
		return nil, err
	}

	tx.Status = models.TxStatusCompleted
	if err := s.txns.UpdateStatusTx(ctx, dbTx, tx.ID, models.TxStatusCompleted); err != nil {
		return nil, err
	}

	if err := dbTx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	s.logAudit(ctx, tx)
	s.log.Info("transfer completed",
		slog.String("from", fromUserID.String()),
		slog.String("to", toUserID.String()),
		slog.String("amount", amount.String()))
	return tx, nil
}

func (s *TransactionService) SubmitAsync(tx *models.Transaction) bool {
	return s.wp.Submit(worker.Task{Transaction: tx})
}

func (s *TransactionService) Rollback(ctx context.Context, txID uuid.UUID) error {
	tx, err := s.txns.GetByID(ctx, txID)
	if err != nil {
		return fmt.Errorf("get transaction: %w", err)
	}

	if err := tx.Transition(models.TxStatusRolledBack); err != nil {
		return err
	}

	dbTx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer dbTx.Rollback(ctx)

	switch tx.Type {
	case models.TxTypeCredit:
		bal, err := s.balances.GetBalanceForUpdate(ctx, dbTx, tx.ToUserID)
		if err != nil {
			return err
		}
		if err := s.balances.UpdateBalanceInTx(ctx, dbTx, tx.ToUserID, bal.Amount.Sub(tx.Amount)); err != nil {
			return err
		}

	case models.TxTypeDebit:
		bal, err := s.balances.GetBalanceForUpdate(ctx, dbTx, tx.FromUserID)
		if err != nil {
			return err
		}
		if err := s.balances.UpdateBalanceInTx(ctx, dbTx, tx.FromUserID, bal.Amount.Add(tx.Amount)); err != nil {
			return err
		}

	case models.TxTypeTransfer:
		senderBal, err := s.balances.GetBalanceForUpdate(ctx, dbTx, tx.FromUserID)
		if err != nil {
			return err
		}
		receiverBal, err := s.balances.GetBalanceForUpdate(ctx, dbTx, tx.ToUserID)
		if err != nil {
			return err
		}
		if err := s.balances.UpdateBalanceInTx(ctx, dbTx, tx.FromUserID, senderBal.Amount.Add(tx.Amount)); err != nil {
			return err
		}
		if err := s.balances.UpdateBalanceInTx(ctx, dbTx, tx.ToUserID, receiverBal.Amount.Sub(tx.Amount)); err != nil {
			return err
		}
	}

	if err := s.txns.UpdateStatusTx(ctx, dbTx, tx.ID, models.TxStatusRolledBack); err != nil {
		return err
	}

	if err := dbTx.Commit(ctx); err != nil {
		return fmt.Errorf("commit rollback: %w", err)
	}

	details, _ := json.Marshal(map[string]string{"original_tx": txID.String()})
	_ = s.audit.Create(ctx, &models.AuditLog{
		ID:         uuid.New(),
		EntityType: models.EntityTransaction,
		EntityID:   txID,
		Action:     models.ActionRollback,
		Details:    details,
		CreatedAt:  time.Now(),
	})

	s.log.Info("transaction rolled back", slog.String("tx_id", txID.String()))
	return nil
}

func (s *TransactionService) ProcessTransaction(ctx context.Context, tx *models.Transaction) error {
	switch tx.Type {
	case models.TxTypeTransfer:
		_, err := s.Transfer(ctx, tx.FromUserID, tx.ToUserID, tx.Amount)
		return err
	case models.TxTypeCredit:
		_, err := s.Credit(ctx, tx.ToUserID, tx.Amount)
		return err
	case models.TxTypeDebit:
		_, err := s.Debit(ctx, tx.FromUserID, tx.Amount)
		return err
	default:
		return fmt.Errorf("unknown transaction type: %s", tx.Type)
	}
}

func (s *TransactionService) logAudit(ctx context.Context, tx *models.Transaction) {
	details, _ := json.Marshal(map[string]string{
		"type":   tx.Type,
		"amount": tx.Amount.String(),
		"status": tx.Status,
		"from":   tx.FromUserID.String(),
		"to":     tx.ToUserID.String(),
	})
	_ = s.audit.Create(ctx, &models.AuditLog{
		ID:         uuid.New(),
		EntityType: models.EntityTransaction,
		EntityID:   tx.ID,
		Action:     models.ActionCreate,
		Details:    details,
		CreatedAt:  time.Now(),
	})
}
