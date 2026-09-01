package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/forgeflow/forgeflow/internal/analytics"
	"github.com/forgeflow/forgeflow/internal/api"
	"github.com/forgeflow/forgeflow/internal/audit"
	"github.com/forgeflow/forgeflow/internal/auth"
	"github.com/forgeflow/forgeflow/internal/config"
	"github.com/forgeflow/forgeflow/internal/database"
	"github.com/forgeflow/forgeflow/internal/events"
	"github.com/forgeflow/forgeflow/internal/jobs"
	"github.com/forgeflow/forgeflow/internal/projects"
	"github.com/forgeflow/forgeflow/internal/repositories"
	"github.com/forgeflow/forgeflow/internal/tasks"
	"github.com/forgeflow/forgeflow/internal/workflows"
	"github.com/forgeflow/forgeflow/pkg/logger"
)

func main(){
	cfg,err:=config.Load();if err!=nil{panic(err)};log:=logger.New(cfg.LogLevel);ctx,stop:=signal.NotifyContext(context.Background(),os.Interrupt,syscall.SIGTERM);defer stop()
	db,err:=database.Open(ctx,cfg.DatabaseURL);if err!=nil{log.Error("database unavailable","error",err);os.Exit(1)};defer db.Close()
	queue,err:=jobs.NewRedisQueue(cfg.RedisURL);if err!=nil{log.Error("queue configuration invalid","error",err);os.Exit(1)};defer queue.Close()
	bus:=events.NewBus();unsubscribe:=audit.Subscribe(bus,db.Pool);defer unsubscribe()
	authService:=auth.NewService(database.NewAuthRepository(db));projectRepository:=database.NewProjectRepository(db);projectService:=projects.NewService(projectRepository,bus);taskService:=tasks.NewService(database.NewTaskRepository(db),projectRepository);workflowService:=workflows.NewService(database.NewWorkflowRepository(db),projectRepository,queue);repositoryService:=repositories.NewService(database.NewRepositoryRepository(db),projectRepository,cfg.WorkspaceRoot);insightService:=analytics.NewService(database.NewInsightsRepository(db),projectRepository)
	server:=api.NewServer(api.Dependencies{Auth:authService,Projects:projectService,Tasks:taskService,Workflows:workflowService,Repositories:repositoryService,Insights:insightService,Database:db,Queue:queue,Logger:log,AllowedOrigins:cfg.AllowedOrigins});server.Addr=cfg.HTTPAddr
	go func(){<-ctx.Done();shutdownCtx,cancel:=context.WithTimeout(context.Background(),10*time.Second);defer cancel();_ = server.Shutdown(shutdownCtx)}()
	log.Info("API listening","address",cfg.HTTPAddr);if err:=server.ListenAndServe();err!=nil&&!errors.Is(err,http.ErrServerClosed){log.Error("API stopped unexpectedly","error",err);os.Exit(1)}
}
