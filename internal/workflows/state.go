package workflows

import (
	"github.com/forgeflow/forgeflow/internal/domain"
	"github.com/forgeflow/forgeflow/pkg/models"
)

var transitions = map[models.RunStatus]map[models.RunStatus]bool{
	models.RunQueued:  {models.RunRunning: true, models.RunCanceled: true},
	models.RunRunning: {models.RunSucceeded: true, models.RunFailed: true, models.RunCanceled: true, models.RunTimedOut: true},
}

func CanTransition(from, to models.RunStatus) bool { return transitions[from][to] }

func Transition(run *models.WorkflowRun, to models.RunStatus) error {
	if !CanTransition(run.Status, to) {
		return &domain.TransitionError{Entity: "workflow run", From: string(run.Status), To: string(to)}
	}
	run.Status = to
	return nil
}

