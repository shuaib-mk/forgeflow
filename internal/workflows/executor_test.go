package workflows

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/forgeflow/forgeflow/pkg/models"
)

type fakeRunner struct {
	calls     map[string]int
	failUntil map[string]int
}

func (r *fakeRunner) Run(_ context.Context, _ string, step models.WorkflowStep, output io.Writer) error {
	r.calls[step.ID]++
	_, _ = io.WriteString(output, step.ID+"\n")
	if r.calls[step.ID] <= r.failUntil[step.ID] {
		return errors.New("expected failure")
	}
	return nil
}

func TestExecutorRetriesAndPropagatesDependencies(t *testing.T) {
	t.Parallel()
	runner := &fakeRunner{calls: map[string]int{}, failUntil: map[string]int{"lint": 1, "test": 2}}
	executor := Executor{Runner: runner, Backoff: func(int) time.Duration { return 0 }}
	result := executor.Execute(context.Background(), Definition{Name: "ci", Steps: []models.WorkflowStep{
		{ID: "lint", Command: "lint", Retries: 1, Timeout: time.Second},
		{ID: "test", Command: "test", DependsOn: []string{"lint"}, Retries: 0, Timeout: time.Second},
		{ID: "build", Command: "build", DependsOn: []string{"test"}, Timeout: time.Second},
	}}, ".")
	if result.Status != models.RunFailed {
		t.Fatalf("status = %s", result.Status)
	}
	if runner.calls["lint"] != 2 {
		t.Fatalf("lint calls = %d", runner.calls["lint"])
	}
	if result.Steps[1].Status != models.RunFailed {
		t.Fatalf("test status = %s", result.Steps[1].Status)
	}
	if result.Steps[2].Status != models.RunCanceled {
		t.Fatalf("build status = %s", result.Steps[2].Status)
	}
}

func TestRunStateMachine(t *testing.T) {
	t.Parallel()
	run := models.WorkflowRun{Status: models.RunQueued}
	if err := Transition(&run, models.RunRunning); err != nil {
		t.Fatal(err)
	}
	if err := Transition(&run, models.RunQueued); err == nil {
		t.Fatal("expected invalid transition")
	}
}
