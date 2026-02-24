package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BilalGunden-Insider/go-backend/internal/api"
	"github.com/BilalGunden-Insider/go-backend/internal/config"
	"github.com/BilalGunden-Insider/go-backend/internal/database"
	"github.com/BilalGunden-Insider/go-backend/internal/logger"
	"github.com/BilalGunden-Insider/go-backend/internal/repository/postgres"
	"github.com/BilalGunden-Insider/go-backend/internal/service"
	"github.com/BilalGunden-Insider/go-backend/internal/worker"
)

func main() {
	cfg := config.Load()

	log := logger.Setup(cfg)
	log.Info("starting server", slog.String("environment", cfg.Environment))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info("running database migrations")
	if err := database.RunMigrations(cfg.DatabaseURL, "migrations"); err != nil {
		log.Error("failed to run migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}
	log.Info("migrations completed")

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()
	log.Info("connected to database")

	userRepo := postgres.NewUserRepository(pool)
	txnRepo := postgres.NewTransactionRepository(pool)
	balanceRepo := postgres.NewBalanceRepository(pool)
	auditRepo := postgres.NewAuditLogRepository(pool)

	balanceSvc := service.NewBalanceService(balanceRepo, auditRepo, log)
	userSvc := service.NewUserService(userRepo, balanceRepo, auditRepo, log)

	if err := balanceSvc.WarmCache(ctx); err != nil {
		log.Error("failed to warm balance cache", slog.String("error", err.Error()))
		os.Exit(1)
	}

	txnSvc := service.NewTransactionService(txnRepo, auditRepo, balanceSvc, pool, log)

	wp := worker.NewPool(4, 100, txnSvc.ProcessTransaction, log)
	wp.Start(ctx)
	txnSvc.SetWorkerPool(wp)

	srv := api.NewServer(cfg, userSvc, txnSvc, balanceSvc, txnRepo, log)

	go func() {
		log.Info("http server listening", slog.String("port", cfg.Port))
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Info("shutting down", slog.String("signal", sig.String()))

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("http server shutdown error", slog.String("error", err.Error()))
	}

	wp.Stop()
	cancel()
	pool.Close()
	log.Info("shutdown complete")
}
