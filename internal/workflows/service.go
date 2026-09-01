package workflows

import (
	"context"
	"time"

	"github.com/forgeflow/forgeflow/internal/authorization"
	"github.com/forgeflow/forgeflow/internal/database"
	"github.com/forgeflow/forgeflow/pkg/models"
	"github.com/google/uuid"
)

type WorkflowRepository interface {
	Create(context.Context, models.Workflow) error
	Get(context.Context, models.ID) (models.Workflow, error)
	CreateRun(context.Context, models.WorkflowRun) error
	GetRun(context.Context, models.ID) (models.WorkflowRun, error)
	Logs(context.Context, models.ID, int) ([]string, error)
}
type ProjectAccess interface {
	Get(context.Context, models.ID) (models.Project, error)
	Role(context.Context, models.ID, models.ID) (models.Role, error)
	List(context.Context, database.ProjectFilter) ([]models.Project, int, error)
	Create(context.Context, models.Project) error
}
type Queue interface {
	Enqueue(context.Context, models.ID) error
}
type Service struct {
	repository WorkflowRepository
	projects   ProjectAccess
	queue      Queue
	now        func() time.Time
}

func NewService(repository WorkflowRepository, projects ProjectAccess, queue Queue) *Service {
	return &Service{repository: repository, projects: projects, queue: queue, now: time.Now}
}

func (s *Service) Create(ctx context.Context, actor models.User, projectID models.ID, definition Definition) (models.Workflow, error) {
	if err := Validate(&definition); err != nil {
		return models.Workflow{}, err
	}
	if err := s.authorize(ctx, actor.ID, projectID, authorization.Write); err != nil {
		return models.Workflow{}, err
	}
	now := s.now().UTC()
	workflow := models.Workflow{ID: models.ID(uuid.NewString()), ProjectID: projectID, Name: definition.Name, Version: 1, Steps: definition.Steps, CreatedAt: now, UpdatedAt: now}
	if err := s.repository.Create(ctx, workflow); err != nil {
		return models.Workflow{}, err
	}
	return workflow, nil
}
func (s *Service) Run(ctx context.Context, actor models.User, workflowID models.ID) (models.WorkflowRun, error) {
	workflow, err := s.repository.Get(ctx, workflowID)
	if err != nil {
		return models.WorkflowRun{}, err
	}
	if err := s.authorize(ctx, actor.ID, workflow.ProjectID, authorization.Write); err != nil {
		return models.WorkflowRun{}, err
	}
	run := models.WorkflowRun{ID: models.ID(uuid.NewString()), WorkflowID: workflow.ID, ProjectID: workflow.ProjectID, Status: models.RunQueued, TriggeredBy: actor.ID, CreatedAt: s.now().UTC()}
	if err := s.repository.CreateRun(ctx, run); err != nil {
		return models.WorkflowRun{}, err
	}
	if err := s.queue.Enqueue(ctx, run.ID); err != nil {
		return models.WorkflowRun{}, err
	}
	return run, nil
}
func (s *Service) GetRun(ctx context.Context, actor models.User, id models.ID) (models.WorkflowRun, error) {
	run, err := s.repository.GetRun(ctx, id)
	if err != nil {
		return models.WorkflowRun{}, err
	}
	if err := s.authorize(ctx, actor.ID, run.ProjectID, authorization.Read); err != nil {
		return models.WorkflowRun{}, err
	}
	return run, nil
}
func (s *Service) Logs(ctx context.Context, actor models.User, id models.ID, after int) ([]string, error) {
	run, err := s.GetRun(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	return s.repository.Logs(ctx, run.ID, after)
}
func (s *Service) authorize(ctx context.Context, userID, projectID models.ID, permission authorization.Permission) error {
	project, err := s.projects.Get(ctx, projectID)
	if err != nil {
		return err
	}
	role, err := s.projects.Role(ctx, project.OrganizationID, userID)
	if err != nil {
		return err
	}
	return authorization.Require(role, permission)
}
