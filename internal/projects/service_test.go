package projects

import (
	"context"
	"testing"

	"github.com/forgeflow/forgeflow/internal/database"
	"github.com/forgeflow/forgeflow/internal/domain"
	"github.com/forgeflow/forgeflow/pkg/models"
)

type projectRepositoryFake struct {
	role    models.Role
	created models.Project
	items   []models.Project
}

func (r *projectRepositoryFake) Create(_ context.Context, project models.Project) error {
	r.created = project
	r.items = append(r.items, project)
	return nil
}
func (r *projectRepositoryFake) Get(_ context.Context, id models.ID) (models.Project, error) {
	for _, project := range r.items {
		if project.ID == id {
			return project, nil
		}
	}
	return models.Project{}, domain.ErrNotFound
}
func (r *projectRepositoryFake) List(_ context.Context, filter database.ProjectFilter) ([]models.Project, int, error) {
	return r.items, len(r.items), nil
}
func (r *projectRepositoryFake) Role(context.Context, models.ID, models.ID) (models.Role, error) {
	if r.role == "" {
		return "", domain.ErrForbidden
	}
	return r.role, nil
}

func TestCreateGetAndListProject(t *testing.T) {
	t.Parallel()
	repository := &projectRepositoryFake{role: models.RoleDeveloper}
	service := NewService(repository, nil)
	actor := models.User{ID: "user-id"}
	project, err := service.Create(context.Background(), actor, CreateInput{OrganizationID: "org-id", Name: " Product ", Slug: "product", Description: " Delivery "})
	if err != nil {
		t.Fatal(err)
	}
	if project.Name != "Product" || project.Description != "Delivery" || project.ID == "" {
		t.Fatalf("project=%+v", project)
	}
	got, err := service.Get(context.Background(), actor, project.ID)
	if err != nil || got.ID != project.ID {
		t.Fatalf("get=%+v err=%v", got, err)
	}
	page, err := service.List(context.Background(), actor, ListInput{OrganizationID: "org-id", Page: -1, PageSize: 500})
	if err != nil || len(page.Items) != 1 || page.Page != 1 || page.PageSize != 100 {
		t.Fatalf("page=%+v err=%v", page, err)
	}
}

func TestCreateProjectValidatesInputAndRole(t *testing.T) {
	t.Parallel()
	actor := models.User{ID: "user-id"}
	if _, err := NewService(&projectRepositoryFake{}, nil).Create(context.Background(), actor, CreateInput{OrganizationID: "org-id", Name: "Product", Slug: "product"}); err != domain.ErrForbidden {
		t.Fatalf("role error=%v", err)
	}
	service := NewService(&projectRepositoryFake{role: models.RoleOwner}, nil)
	for _, input := range []CreateInput{
		{OrganizationID: "org-id", Slug: "product"},
		{OrganizationID: "org-id", Name: "Product", Slug: "Not Valid"},
		{OrganizationID: "org-id", Name: "Product", Slug: "product", Description: string(make([]byte, 2001))},
	} {
		if _, err := service.Create(context.Background(), actor, input); err == nil {
			t.Fatalf("input=%+v expected validation error", input)
		}
	}
}
