package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bilal/backend_path/internal/config"
	"github.com/bilal/backend_path/internal/database"
	"github.com/bilal/backend_path/internal/logger"
	"github.com/bilal/backend_path/internal/repository/postgres"
	"github.com/bilal/backend_path/internal/service"
	"github.com/bilal/backend_path/internal/worker"
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

	_ = userSvc
	_ = txnSvc

	log.Info("all services initialized")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Info("shutting down", slog.String("signal", sig.String()))

	wp.Stop()
	cancel()
	pool.Close()
	log.Info("shutdown complete")
}
