package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/forgeflow/forgeflow/internal/database"
	"github.com/forgeflow/forgeflow/internal/workflows"
	"github.com/forgeflow/forgeflow/pkg/models"
)

type Worker struct {
	Queue         Queue
	Repository    *database.WorkflowRepository
	Executor      workflows.Executor
	WorkspaceRoot string
	Concurrency   int
	Logger        *slog.Logger
}

func (w *Worker) Run(ctx context.Context) error {
	if w.Concurrency < 1 {
		return errors.New("worker concurrency must be positive")
	}
	var group sync.WaitGroup
	errorsChannel := make(chan error, w.Concurrency)
	for index := 0; index < w.Concurrency; index++ {
		group.Add(1)
		go func(workerID int) {
			defer group.Done()
			if err := w.loop(ctx, workerID); err != nil && !errors.Is(err, context.Canceled) {
				errorsChannel <- err
			}
		}(index + 1)
	}
	done := make(chan struct{})
	go func() { group.Wait(); close(done) }()
	select {
	case <-ctx.Done():
		<-done
		return nil
	case err := <-errorsChannel:
		return err
	case <-done:
		return nil
	}
}

func (w *Worker) loop(ctx context.Context, workerID int) error {
	for {
		runID, err := w.Queue.Dequeue(ctx, 2*time.Second)
		if errors.Is(err, context.DeadlineExceeded) {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			continue
		}
		if err != nil {
			return err
		}
		if err := w.process(ctx, runID); err != nil {
			w.Logger.ErrorContext(ctx, "workflow job failed", "worker", workerID, "run_id", runID, "error", err)
		}
	}
}

func (w *Worker) process(ctx context.Context, runID models.ID) error {
	run, claimed, err := w.Repository.ClaimRun(ctx, runID)
	if err != nil || !claimed {
		return err
	}
	workflow, err := w.Repository.Get(ctx, run.WorkflowID)
	if err != nil {
		_ = w.Repository.CompleteRun(ctx, run.ID, models.RunFailed, err.Error())
		return err
	}
	definition := workflows.Definition{Name: workflow.Name, Steps: workflow.Steps}
	executionDirectory, err := w.Repository.ExecutionDirectory(ctx, run.ProjectID)
	if err != nil {
		_ = w.Repository.CompleteRun(ctx, run.ID, models.RunFailed, err.Error())
		return err
	}
	if executionDirectory == "" {
		executionDirectory = w.WorkspaceRoot
	}
	result := w.Executor.Execute(ctx, definition, executionDirectory)
	for sequence, step := range result.Steps {
		if step.Log != "" {
			_ = w.Repository.AppendLog(ctx, run.ID, step.StepID, sequence, step.Log)
		}
	}
	message := ""
	if result.Status != models.RunSucceeded {
		message = "one or more workflow steps failed"
	}
	if err := w.Repository.CompleteRun(ctx, run.ID, result.Status, message); err != nil {
		return fmt.Errorf("complete run: %w", err)
	}
	return nil
}
