package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/BilalGunden-Insider/go-backend/internal/cache"
	"github.com/BilalGunden-Insider/go-backend/internal/circuitbreaker"
	"github.com/BilalGunden-Insider/go-backend/internal/models"
	"github.com/BilalGunden-Insider/go-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type BalanceService struct {
	balances repository.BalanceRepository
	audit    repository.AuditLogRepository
	cache    cache.Cache
	cb       *circuitbreaker.CircuitBreaker
	log      *slog.Logger
}

func NewBalanceService(
	balances repository.BalanceRepository,
	audit repository.AuditLogRepository,
	cache cache.Cache,
	cb *circuitbreaker.CircuitBreaker,
	log *slog.Logger,
) *BalanceService {
	return &BalanceService{
		balances: balances,
		audit:    audit,
		cache:    cache,
		cb:       cb,
		log:      log,
	}
}

func (s *BalanceService) WarmCache(ctx context.Context) error {
	var all []*models.Balance
	err := s.cb.Execute(func() error {
		var e error
		all, e = s.balances.GetAll(ctx)
		return e
	})
	if err != nil {
		if errors.Is(err, circuitbreaker.ErrCircuitOpen) {
			s.log.Warn("circuit open during cache warm-up, starting with empty cache")
			return nil
		}
		return fmt.Errorf("warm cache: %w", err)
	}

	batch := make(map[uuid.UUID]decimal.Decimal, len(all))
	for _, b := range all {
		batch[b.UserID] = b.Amount
	}

	if err := s.cache.SetBalanceBatch(ctx, batch); err != nil {
		return fmt.Errorf("warm cache batch set: %w", err)
	}

	s.log.Info("balance cache warmed", slog.Int("entries", len(all)))
	return nil
}

func (s *BalanceService) GetBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	if amount, ok, err := s.cache.GetBalance(ctx, userID); err == nil && ok {
		return amount, nil
	}

	var bal *models.Balance
	err := s.cb.Execute(func() error {
		var e error
		bal, e = s.balances.GetByUserID(ctx, userID)
		return e
	})
	if err != nil {
		if errors.Is(err, circuitbreaker.ErrCircuitOpen) {
			s.log.Warn("circuit open, cannot fetch balance", slog.String("user_id", userID.String()))
			return decimal.Zero, fmt.Errorf("service temporarily unavailable")
		}
		return decimal.Zero, err
	}

	_ = s.cache.SetBalance(ctx, userID, bal.Amount)
	return bal.Amount, nil
}

func (s *BalanceService) UpdateBalanceInTx(ctx context.Context, dbTx pgx.Tx, userID uuid.UUID, newAmount decimal.Decimal) error {
	if err := s.balances.UpdateAmount(ctx, dbTx, userID, newAmount); err != nil {
		return err
	}

	_ = s.cache.SetBalance(ctx, userID, newAmount)
	return nil
}

func (s *BalanceService) GetBalanceForUpdate(ctx context.Context, dbTx pgx.Tx, userID uuid.UUID) (*models.Balance, error) {
	return s.balances.GetByUserIDForUpdate(ctx, dbTx, userID)
}

func (s *BalanceService) SetCache(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) {
	_ = s.cache.SetBalance(ctx, userID, amount)
}
