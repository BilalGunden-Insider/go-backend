package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/BilalGunden-Insider/go-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Limiter struct {
	limits repository.TransactionLimitRepository
	txns   repository.TransactionRepository
	log    *slog.Logger
}

func NewLimiter(
	limits repository.TransactionLimitRepository,
	txns repository.TransactionRepository,
	log *slog.Logger,
) *Limiter {
	return &Limiter{limits: limits, txns: txns, log: log}
}

func (l *Limiter) CheckLimits(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) error {
	limit, err := l.limits.GetByUserID(ctx, userID)
	if err != nil {
		l.log.Warn("failed to get limits, skipping check", slog.String("error", err.Error()))
		return nil
	}

	if amount.GreaterThan(limit.MaxPerTransaction) {
		return fmt.Errorf("amount %s exceeds per-transaction limit of %s", amount, limit.MaxPerTransaction)
	}

	dailyTotal, err := l.txns.GetDailyTotal(ctx, userID, time.Now())
	if err != nil {
		l.log.Warn("failed to get daily total, skipping check", slog.String("error", err.Error()))
		return nil
	}

	if dailyTotal.Add(amount).GreaterThan(limit.MaxDailyAmount) {
		return fmt.Errorf("daily limit exceeded: used %s of %s, requested %s", dailyTotal, limit.MaxDailyAmount, amount)
	}

	return nil
}
