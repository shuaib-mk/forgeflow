package tasks

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/forgeflow/forgeflow/internal/authorization"
	"github.com/forgeflow/forgeflow/internal/database"
	"github.com/forgeflow/forgeflow/internal/domain"
	"github.com/forgeflow/forgeflow/pkg/models"
	"github.com/forgeflow/forgeflow/pkg/validation"
	"github.com/google/uuid"
)

type TaskRepository interface {
	Create(context.Context, models.Task) error
	List(context.Context, models.ID, models.TaskStatus, int, int) ([]models.Task, int, error)
	UpdateStatus(context.Context, models.ID, models.TaskStatus) (models.Task, error)
}
type ProjectRepository interface {
	Get(context.Context, models.ID) (models.Project, error)
	Role(context.Context, models.ID, models.ID) (models.Role, error)
	List(context.Context, database.ProjectFilter) ([]models.Project, int, error)
	Create(context.Context, models.Project) error
}
type Service struct {
	tasks    TaskRepository
	projects ProjectRepository
	now      func() time.Time
}

func NewService(tasks TaskRepository, projects ProjectRepository) *Service {
	return &Service{tasks: tasks, projects: projects, now: time.Now}
}

type CreateInput struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	AssigneeID  *models.ID `json:"assigneeId,omitempty"`
}

func (s *Service) Create(ctx context.Context, actor models.User, projectID models.ID, input CreateInput) (models.Task, error) {
	role, err := s.role(ctx, actor.ID, projectID)
	if err != nil {
		return models.Task{}, err
	}
	if err := authorization.Require(role, authorization.Write); err != nil {
		return models.Task{}, err
	}
	if !validation.Required(input.Title, 200) {
		return models.Task{}, domain.Invalid("title", "must contain 1 to 200 characters")
	}
	if len(input.Description) > 10000 {
		return models.Task{}, domain.Invalid("description", "must not exceed 10000 characters")
	}
	now := s.now().UTC()
	task := models.Task{ID: models.ID(uuid.NewString()), ProjectID: projectID, Title: strings.TrimSpace(input.Title), Description: strings.TrimSpace(input.Description), Status: models.TaskOpen, AssigneeID: input.AssigneeID, CreatedBy: actor.ID, CreatedAt: now, UpdatedAt: now}
	if err := s.tasks.Create(ctx, task); err != nil {
		return models.Task{}, err
	}
	return task, nil
}
func (s *Service) List(ctx context.Context, actor models.User, projectID models.ID, status models.TaskStatus, page, pageSize int) (models.Page[models.Task], error) {
	role, err := s.role(ctx, actor.ID, projectID)
	if err != nil {
		return models.Page[models.Task]{}, err
	}
	if err := authorization.Require(role, authorization.Read); err != nil {
		return models.Page[models.Task]{}, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	items, total, err := s.tasks.List(ctx, projectID, status, pageSize, (page-1)*pageSize)
	if err != nil {
		return models.Page[models.Task]{}, err
	}
	return models.Page[models.Task]{Items: items, Page: page, PageSize: pageSize, TotalItems: total, TotalPages: int(math.Ceil(float64(total) / float64(pageSize)))}, nil
}
func (s *Service) UpdateStatus(ctx context.Context, actor models.User, projectID, taskID models.ID, status models.TaskStatus) (models.Task, error) {
	role, err := s.role(ctx, actor.ID, projectID)
	if err != nil {
		return models.Task{}, err
	}
	if err := authorization.Require(role, authorization.Write); err != nil {
		return models.Task{}, err
	}
	if status != models.TaskOpen && status != models.TaskInProgress && status != models.TaskDone && status != models.TaskCanceled {
		return models.Task{}, domain.Invalid("status", "is not a valid task status")
	}
	task, err := s.tasks.UpdateStatus(ctx, taskID, status)
	if err != nil {
		return models.Task{}, err
	}
	if task.ProjectID != projectID {
		return models.Task{}, domain.ErrNotFound
	}
	return task, nil
}
func (s *Service) role(ctx context.Context, userID, projectID models.ID) (models.Role, error) {
	project, err := s.projects.Get(ctx, projectID)
	if err != nil {
		return "", err
	}
	return s.projects.Role(ctx, project.OrganizationID, userID)
}
