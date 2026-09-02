package workflows

import (
	"context"
	"testing"

	"github.com/forgeflow/forgeflow/internal/database"
	"github.com/forgeflow/forgeflow/internal/domain"
	"github.com/forgeflow/forgeflow/pkg/models"
)

type workflowRepositoryFake struct {
	workflow models.Workflow
	runs     []models.WorkflowRun
	logs     []string
}

func (r *workflowRepositoryFake) Create(_ context.Context, workflow models.Workflow) error {
	r.workflow = workflow
	return nil
}
func (r *workflowRepositoryFake) Get(_ context.Context, id models.ID) (models.Workflow, error) {
	if r.workflow.ID != id {
		return models.Workflow{}, domain.ErrNotFound
	}
	return r.workflow, nil
}
func (r *workflowRepositoryFake) ListWorkflows(context.Context, models.ID) ([]models.Workflow, error) {
	if r.workflow.ID == "" {
		return []models.Workflow{}, nil
	}
	return []models.Workflow{r.workflow}, nil
}
func (r *workflowRepositoryFake) CreateRun(_ context.Context, run models.WorkflowRun) error {
	r.runs = append(r.runs, run)
	return nil
}
func (r *workflowRepositoryFake) GetRun(_ context.Context, id models.ID) (models.WorkflowRun, error) {
	for _, run := range r.runs {
		if run.ID == id {
			return run, nil
		}
	}
	return models.WorkflowRun{}, domain.ErrNotFound
}
func (r *workflowRepositoryFake) ListRuns(context.Context, models.ID, models.ID, int, int) ([]models.WorkflowRun, int, error) {
	return r.runs, len(r.runs), nil
}
func (r *workflowRepositoryFake) Logs(context.Context, models.ID, int) ([]string, error) {
	return r.logs, nil
}

type workflowProjectsFake struct{ role models.Role }

func (p workflowProjectsFake) Get(_ context.Context, id models.ID) (models.Project, error) {
	if id != "project-id" {
		return models.Project{}, domain.ErrNotFound
	}
	return models.Project{ID: id, OrganizationID: "org-id"}, nil
}
func (p workflowProjectsFake) Role(context.Context, models.ID, models.ID) (models.Role, error) {
	if p.role == "" {
		return "", domain.ErrForbidden
	}
	return p.role, nil
}
func (workflowProjectsFake) List(context.Context, database.ProjectFilter) ([]models.Project, int, error) {
	return nil, 0, nil
}
func (workflowProjectsFake) Create(context.Context, models.Project) error { return nil }

type queueFake struct{ runID models.ID }

func (q *queueFake) Enqueue(_ context.Context, id models.ID) error { q.runID = id; return nil }

func TestWorkflowLifecycle(t *testing.T) {
	t.Parallel()
	repository := &workflowRepositoryFake{logs: []string{"done\n"}}
	queue := &queueFake{}
	service := NewService(repository, workflowProjectsFake{role: models.RoleDeveloper}, queue)
	actor := models.User{ID: "user-id"}
	definition := Definition{Name: "check", Steps: []models.WorkflowStep{{ID: "test", Command: "git", Args: []string{"--version"}}}}
	workflow, err := service.Create(context.Background(), actor, "project-id", definition)
	if err != nil || workflow.ID == "" || workflow.Steps[0].Timeout == 0 {
		t.Fatalf("workflow=%+v err=%v", workflow, err)
	}
	items, err := service.List(context.Background(), actor, "project-id")
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%+v err=%v", items, err)
	}
	run, err := service.Run(context.Background(), actor, workflow.ID)
	if err != nil || run.Status != models.RunQueued || queue.runID != run.ID {
		t.Fatalf("run=%+v queued=%s err=%v", run, queue.runID, err)
	}
	got, err := service.GetRun(context.Background(), actor, run.ID)
	if err != nil || got.ID != run.ID {
		t.Fatalf("get=%+v err=%v", got, err)
	}
	logs, err := service.Logs(context.Background(), actor, run.ID, -1)
	if err != nil || len(logs) != 1 {
		t.Fatalf("logs=%v err=%v", logs, err)
	}
	page, err := service.ListRuns(context.Background(), actor, "org-id", "project-id", 0, 500)
	if err != nil || len(page.Items) != 1 || page.Page != 1 || page.PageSize != 100 {
		t.Fatalf("page=%+v err=%v", page, err)
	}
}

func TestWorkflowAuthorizationAndProjectScope(t *testing.T) {
	t.Parallel()
	definition := Definition{Name: "check", Steps: []models.WorkflowStep{{ID: "test", Command: "git"}}}
	service := NewService(&workflowRepositoryFake{}, workflowProjectsFake{}, &queueFake{})
	if _, err := service.Create(context.Background(), models.User{ID: "user-id"}, "project-id", definition); err != domain.ErrForbidden {
		t.Fatalf("create error=%v", err)
	}
	allowed := NewService(&workflowRepositoryFake{}, workflowProjectsFake{role: models.RoleDeveloper}, &queueFake{})
	if _, err := allowed.ListRuns(context.Background(), models.User{ID: "user-id"}, "other-org", "project-id", 1, 20); err != domain.ErrNotFound {
		t.Fatalf("scope error=%v", err)
	}
}
