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

	"task-management/internal/config"
	"task-management/internal/delivery"
	"task-management/internal/repository/postgres"
	"task-management/internal/service"
	"task-management/internal/worker"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := postgres.Open(ctx, cfg.Database)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("failed to close database", "error", err)
		}
	}()

	userRepo := postgres.NewUserRepository(db)
	taskRepo := postgres.NewTaskRepository(db)

	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.TokenTTL)
	taskSvc := service.NewTaskService(taskRepo, userRepo)

	router := delivery.NewRouter(authSvc, taskSvc)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	deadlineWorker := worker.NewDeadlineWorker(taskRepo, cfg.DeadlineWorkerInterval)
	workerDone := make(chan struct{})
	go func() {
		deadlineWorker.Run(ctx)
		close(workerDone)
	}()

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("starting server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		serverErr <- nil
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case err := <-serverErr:
		if err != nil {
			slog.Error("server failed to start", "error", err)
		}
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful server shutdown failed", "error", err)
		_ = srv.Close()
	}

	select {
	case <-workerDone:
		slog.Info("deadline worker stopped successfully")
	case <-shutdownCtx.Done():
		slog.Error("worker shutdown timed out")
	}

	slog.Info("service stopped")
}
