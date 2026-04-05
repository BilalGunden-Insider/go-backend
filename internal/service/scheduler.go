package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/BilalGunden-Insider/go-backend/internal/models"
	"github.com/BilalGunden-Insider/go-backend/internal/repository"
)

type Scheduler struct {
	repo   repository.ScheduledTransactionRepository
	txnSvc *TransactionService
	log    *slog.Logger
	done   chan struct{}
}

func NewScheduler(
	repo repository.ScheduledTransactionRepository,
	txnSvc *TransactionService,
	log *slog.Logger,
) *Scheduler {
	return &Scheduler{
		repo:   repo,
		txnSvc: txnSvc,
		log:    log,
		done:   make(chan struct{}),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		defer close(s.done)

		s.log.Info("scheduler started", slog.String("interval", "30s"))

		for {
			select {
			case <-ctx.Done():
				s.log.Info("scheduler stopped")
				return
			case <-ticker.C:
				s.tick(ctx)
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	<-s.done
}

func (s *Scheduler) tick(ctx context.Context) {
	due, err := s.repo.GetDue(ctx, time.Now(), 50)
	if err != nil {
		s.log.Error("scheduler: failed to get due transactions", slog.String("error", err.Error()))
		return
	}

	for _, st := range due {
		s.executeOne(ctx, st)
	}

	if len(due) > 0 {
		s.log.Info("scheduler tick", slog.Int("processed", len(due)))
	}
}

func (s *Scheduler) executeOne(ctx context.Context, st *models.ScheduledTransaction) {
	var err error

	switch st.Type {
	case models.TxTypeCredit:
		_, err = s.txnSvc.Credit(ctx, st.ToUserID, st.Amount)
	case models.TxTypeDebit:
		_, err = s.txnSvc.Debit(ctx, st.FromUserID, st.Amount)
	case models.TxTypeTransfer:
		_, err = s.txnSvc.Transfer(ctx, st.FromUserID, st.ToUserID, st.Amount)
	default:
		err = s.repo.UpdateStatus(ctx, st.ID, models.SchedStatusFailed, "unknown type: "+st.Type)
		return
	}

	if err != nil {
		s.log.Error("scheduler: transaction failed",
			slog.String("id", st.ID.String()),
			slog.String("error", err.Error()))
		_ = s.repo.UpdateStatus(ctx, st.ID, models.SchedStatusFailed, err.Error())
		return
	}

	_ = s.repo.SetExecuted(ctx, st.ID)
}
