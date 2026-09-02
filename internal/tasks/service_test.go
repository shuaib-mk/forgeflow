package tasks

import (
	"context"
	"testing"

	"github.com/forgeflow/forgeflow/internal/database"
	"github.com/forgeflow/forgeflow/internal/domain"
	"github.com/forgeflow/forgeflow/pkg/models"
)

type taskStoreFake struct{ items []models.Task }

func (s *taskStoreFake) Create(_ context.Context, task models.Task) error {
	s.items = append(s.items, task)
	return nil
}
func (s *taskStoreFake) List(context.Context, models.ID, models.TaskStatus, int, int) ([]models.Task, int, error) {
	return s.items, len(s.items), nil
}
func (s *taskStoreFake) UpdateStatus(_ context.Context, projectID, taskID models.ID, status models.TaskStatus) (models.Task, error) {
	for index := range s.items {
		if s.items[index].ID == taskID && s.items[index].ProjectID == projectID {
			s.items[index].Status = status
			return s.items[index], nil
		}
	}
	return models.Task{}, domain.ErrNotFound
}

type taskProjectsFake struct{ role models.Role }

func (p taskProjectsFake) Get(context.Context, models.ID) (models.Project, error) {
	return models.Project{ID: "project-id", OrganizationID: "org-id"}, nil
}
func (p taskProjectsFake) Role(context.Context, models.ID, models.ID) (models.Role, error) {
	return p.role, nil
}
func (p taskProjectsFake) List(context.Context, database.ProjectFilter) ([]models.Project, int, error) {
	return nil, 0, nil
}
func (p taskProjectsFake) Create(context.Context, models.Project) error { return nil }

func TestTaskLifecycle(t *testing.T) {
	t.Parallel()
	store := &taskStoreFake{}
	service := NewService(store, taskProjectsFake{role: models.RoleDeveloper})
	actor := models.User{ID: "user-id"}
	task, err := service.Create(context.Background(), actor, "project-id", CreateInput{Title: " Ship release ", Description: " Verify it "})
	if err != nil || task.Title != "Ship release" || task.Status != models.TaskOpen {
		t.Fatalf("task=%+v err=%v", task, err)
	}
	page, err := service.List(context.Background(), actor, "project-id", "", 0, 1000)
	if err != nil || len(page.Items) != 1 || page.Page != 1 || page.PageSize != 100 {
		t.Fatalf("page=%+v err=%v", page, err)
	}
	updated, err := service.UpdateStatus(context.Background(), actor, "project-id", task.ID, models.TaskDone)
	if err != nil || updated.Status != models.TaskDone {
		t.Fatalf("updated=%+v err=%v", updated, err)
	}
}

func TestTasksRejectInvalidWrites(t *testing.T) {
	t.Parallel()
	actor := models.User{ID: "user-id"}
	viewer := NewService(&taskStoreFake{}, taskProjectsFake{role: models.RoleViewer})
	if _, err := viewer.Create(context.Background(), actor, "project-id", CreateInput{Title: "Task"}); err != domain.ErrForbidden {
		t.Fatalf("viewer create error=%v", err)
	}
	service := NewService(&taskStoreFake{}, taskProjectsFake{role: models.RoleDeveloper})
	if _, err := service.Create(context.Background(), actor, "project-id", CreateInput{}); err == nil {
		t.Fatal("expected title validation error")
	}
	if _, err := service.UpdateStatus(context.Background(), actor, "project-id", "task-id", "invalid"); err == nil {
		t.Fatal("expected status validation error")
	}
}
