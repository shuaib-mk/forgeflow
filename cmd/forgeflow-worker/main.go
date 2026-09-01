package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/forgeflow/forgeflow/internal/config"
	"github.com/forgeflow/forgeflow/internal/database"
	"github.com/forgeflow/forgeflow/internal/jobs"
	"github.com/forgeflow/forgeflow/internal/workflows"
	"github.com/forgeflow/forgeflow/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log := logger.New(cfg.LogLevel)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("database unavailable", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	queue, err := jobs.NewRedisQueue(cfg.RedisURL)
	if err != nil {
		log.Error("queue configuration invalid", "error", err)
		os.Exit(1)
	}
	defer queue.Close()
	worker := jobs.Worker{Queue: queue, Repository: database.NewWorkflowRepository(db), Executor: workflows.Executor{Runner: workflows.LocalCommandRunner{WorkspaceRoot: cfg.WorkspaceRoot}}, WorkspaceRoot: cfg.WorkspaceRoot, Concurrency: cfg.WorkerConcurrency, Logger: log}
	log.Info("worker started", "concurrency", cfg.WorkerConcurrency)
	if err := worker.Run(ctx); err != nil {
		log.Error("worker stopped unexpectedly", "error", err)
		os.Exit(1)
	}
	log.Info("worker stopped")
}
