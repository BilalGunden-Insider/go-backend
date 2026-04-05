package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/BilalGunden-Insider/go-backend/internal/api/handler"
	"github.com/BilalGunden-Insider/go-backend/internal/api/middleware"
	"github.com/BilalGunden-Insider/go-backend/internal/config"
	_ "github.com/BilalGunden-Insider/go-backend/internal/metrics"
	"github.com/BilalGunden-Insider/go-backend/internal/models"
	"github.com/BilalGunden-Insider/go-backend/internal/repository"
	"github.com/BilalGunden-Insider/go-backend/internal/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(
	cfg *config.Config,
	userSvc *service.UserService,
	txnSvc *service.TransactionService,
	balanceSvc *service.BalanceService,
	txnRepo repository.TransactionRepository,
	schedRepo repository.ScheduledTransactionRepository,
	log *slog.Logger,
) *Server {
	mux := http.NewServeMux()

	authH := handler.NewAuthHandler(userSvc, cfg.JWTSecret)
	userH := handler.NewUserHandler(userSvc)
	txnH := handler.NewTransactionHandler(txnSvc, txnRepo)
	balH := handler.NewBalanceHandler(balanceSvc, txnRepo)
	schedH := handler.NewScheduledTransactionHandler(schedRepo)
	batchH := handler.NewBatchHandler(txnSvc)

	base := func(h http.HandlerFunc) http.Handler {
		return middleware.Chain(h,
			middleware.Recover(log),
			middleware.Logger(log),
			middleware.CORS(),
			middleware.RateLimit(100),
		)
	}
	authed := func(h http.HandlerFunc) http.Handler {
		return middleware.Chain(h,
			middleware.Recover(log),
			middleware.Logger(log),
			middleware.CORS(),
			middleware.RateLimit(100),
			middleware.Auth(cfg.JWTSecret),
		)
	}
	admin := func(h http.HandlerFunc) http.Handler {
		return middleware.Chain(h,
			middleware.Recover(log),
			middleware.Logger(log),
			middleware.CORS(),
			middleware.RateLimit(100),
			middleware.Auth(cfg.JWTSecret),
			middleware.RequireRole(models.RoleAdmin),
		)
	}

	mux.Handle("GET /metrics", promhttp.Handler())

	mux.Handle("POST /api/v1/auth/register", base(authH.Register))
	mux.Handle("POST /api/v1/auth/login", base(authH.Login))

	mux.Handle("GET /api/v1/users", admin(userH.ListUsers))
	mux.Handle("GET /api/v1/users/{id}", authed(userH.GetUser))
	mux.Handle("PUT /api/v1/users/{id}", authed(userH.UpdateUser))
	mux.Handle("DELETE /api/v1/users/{id}", admin(userH.DeleteUser))

	mux.Handle("POST /api/v1/transactions/credit", admin(txnH.Credit))
	mux.Handle("POST /api/v1/transactions/debit", admin(txnH.Debit))
	mux.Handle("POST /api/v1/transactions/transfer", authed(txnH.Transfer))
	mux.Handle("GET /api/v1/transactions", authed(txnH.ListTransactions))
	mux.Handle("GET /api/v1/transactions/{id}", authed(txnH.GetTransaction))
	mux.Handle("POST /api/v1/transactions/{id}/rollback", admin(txnH.RollbackTransaction))

	mux.Handle("GET /api/v1/balances/{user_id}", authed(balH.GetBalance))
	mux.Handle("GET /api/v1/balances/{user_id}/at", authed(balH.GetBalanceAt))

	mux.Handle("POST /api/v1/scheduled-transactions", authed(schedH.Schedule))
	mux.Handle("GET /api/v1/scheduled-transactions", authed(schedH.List))
	mux.Handle("GET /api/v1/scheduled-transactions/{id}", authed(schedH.Get))
	mux.Handle("POST /api/v1/scheduled-transactions/{id}/cancel", authed(schedH.Cancel))

	mux.Handle("POST /api/v1/transactions/batch", admin(batchH.ProcessBatch))
	mux.Handle("GET /api/v1/workers/stats", admin(batchH.WorkerStats))

	return &Server{
		httpServer: &http.Server{
			Addr:    ":" + cfg.Port,
			Handler: mux,
		},
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
