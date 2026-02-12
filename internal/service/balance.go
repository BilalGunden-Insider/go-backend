package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/BilalGunden-Insider/go-backend/internal/models"
	"github.com/BilalGunden-Insider/go-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type BalanceService struct {
	balances repository.BalanceRepository
	audit    repository.AuditLogRepository
	log      *slog.Logger

	mu    sync.RWMutex
	cache map[uuid.UUID]decimal.Decimal
}

func NewBalanceService(
	balances repository.BalanceRepository,
	audit repository.AuditLogRepository,
	log *slog.Logger,
) *BalanceService {
	return &BalanceService{
		balances: balances,
		audit:    audit,
		log:      log,
		cache:    make(map[uuid.UUID]decimal.Decimal),
	}
}

func (s *BalanceService) WarmCache(ctx context.Context) error {
	all, err := s.balances.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("warm cache: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, b := range all {
		s.cache[b.UserID] = b.Amount
	}
	s.log.Info("balance cache warmed", slog.Int("entries", len(all)))
	return nil
}

func (s *BalanceService) GetBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	s.mu.RLock()
	if amount, ok := s.cache[userID]; ok {
		s.mu.RUnlock()
		return amount, nil
	}
	s.mu.RUnlock()

	b, err := s.balances.GetByUserID(ctx, userID)
	if err != nil {
		return decimal.Zero, err
	}

	s.mu.Lock()
	s.cache[userID] = b.Amount
	s.mu.Unlock()
	return b.Amount, nil
}

func (s *BalanceService) UpdateBalanceInTx(ctx context.Context, dbTx pgx.Tx, userID uuid.UUID, newAmount decimal.Decimal) error {
	if err := s.balances.UpdateAmount(ctx, dbTx, userID, newAmount); err != nil {
		return err
	}

	s.mu.Lock()
	s.cache[userID] = newAmount
	s.mu.Unlock()
	return nil
}

func (s *BalanceService) GetBalanceForUpdate(ctx context.Context, dbTx pgx.Tx, userID uuid.UUID) (*models.Balance, error) {
	return s.balances.GetByUserIDForUpdate(ctx, dbTx, userID)
}

func (s *BalanceService) SetCache(userID uuid.UUID, amount decimal.Decimal) {
	s.mu.Lock()
	s.cache[userID] = amount
	s.mu.Unlock()
}
