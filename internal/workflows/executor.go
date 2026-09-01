package workflows

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/forgeflow/forgeflow/pkg/models"
)

const maxStepLogBytes = 2 << 20

type StepResult struct {
	StepID    string
	Status    models.RunStatus
	Attempt   int
	StartedAt time.Time
	EndedAt   time.Time
	Log       string
	Err       error
}

type Result struct {
	Status models.RunStatus
	Steps  []StepResult
}

type CommandRunner interface {
	Run(context.Context, string, models.WorkflowStep, io.Writer) error
}

type LocalCommandRunner struct { WorkspaceRoot string }

func (r LocalCommandRunner) Run(ctx context.Context, workingDirectory string, step models.WorkflowStep, output io.Writer) error {
	root, err := filepath.Abs(r.WorkspaceRoot)
	if err != nil { return fmt.Errorf("resolve workspace root: %w", err) }
	workingDirectory, err = filepath.Abs(workingDirectory)
	if err != nil { return fmt.Errorf("resolve working directory: %w", err) }
	relative, err := filepath.Rel(root, workingDirectory)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return errors.New("working directory is outside the configured workspace root")
	}
	command := exec.CommandContext(ctx, step.Command, step.Args...)
	command.Dir = workingDirectory
	command.Stdout = output
	command.Stderr = output
	if err := command.Run(); err != nil { return fmt.Errorf("execute %s: %w", step.Command, err) }
	return nil
}

type Executor struct {
	Runner CommandRunner
	Backoff func(int) time.Duration
}

func (e Executor) Execute(ctx context.Context, definition Definition, workingDirectory string) Result {
	if e.Runner == nil { return Result{Status: models.RunFailed} }
	backoff := e.Backoff
	if backoff == nil { backoff = func(attempt int) time.Duration { return time.Duration(1<<uint(attempt-1)) * time.Second } }
	statuses := make(map[string]models.RunStatus, len(definition.Steps))
	result := Result{Status: models.RunSucceeded}
	for _, step := range definition.Steps {
		blocked := false
		for _, dependency := range step.DependsOn {
			if statuses[dependency] != models.RunSucceeded { blocked = true; break }
		}
		if blocked {
			statuses[step.ID] = models.RunCanceled
			result.Steps = append(result.Steps, StepResult{StepID: step.ID, Status: models.RunCanceled, Err: errors.New("dependency did not succeed")})
			continue
		}
		stepResult := e.executeStep(ctx, workingDirectory, step, backoff)
		statuses[step.ID] = stepResult.Status
		result.Steps = append(result.Steps, stepResult)
		if stepResult.Status != models.RunSucceeded && !step.ContinueOnFail { result.Status = stepResult.Status }
	}
	return result
}

func (e Executor) executeStep(ctx context.Context, directory string, step models.WorkflowStep, backoff func(int) time.Duration) StepResult {
	result := StepResult{StepID: step.ID, Status: models.RunFailed, StartedAt: time.Now().UTC()}
	var log bytes.Buffer
	limited := &limitedWriter{writer: &log, remaining: maxStepLogBytes}
	for attempt := 1; attempt <= step.Retries+1; attempt++ {
		result.Attempt = attempt
		stepCtx, cancel := context.WithTimeout(ctx, step.Timeout)
		err := e.Runner.Run(stepCtx, directory, step, limited)
		contextErr := stepCtx.Err()
		cancel()
		if err == nil {
			result.Status = models.RunSucceeded
			result.Err = nil
			break
		}
		result.Err = err
		if errors.Is(contextErr, context.DeadlineExceeded) { result.Status = models.RunTimedOut } else if errors.Is(contextErr, context.Canceled) { result.Status = models.RunCanceled }
		if attempt <= step.Retries {
			select {
			case <-ctx.Done(): result.Status = models.RunCanceled; result.Err = ctx.Err(); attempt = step.Retries + 1
			case <-time.After(backoff(attempt)):
			}
		}
	}
	result.EndedAt = time.Now().UTC()
	result.Log = log.String()
	return result
}

type limitedWriter struct { writer io.Writer; remaining int }
func (w *limitedWriter) Write(p []byte) (int, error) {
	original := len(p)
	if w.remaining <= 0 { return original, nil }
	if len(p) > w.remaining { p = p[:w.remaining] }
	_, err := w.writer.Write(p)
	w.remaining -= len(p)
	return original, err
}
