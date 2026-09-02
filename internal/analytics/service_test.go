package analytics

import (
	"context"
	"testing"

	"github.com/forgeflow/forgeflow/internal/database"
	"github.com/forgeflow/forgeflow/internal/domain"
	"github.com/forgeflow/forgeflow/pkg/models"
)

type insightStoreFake struct{}

func (insightStoreFake) Analytics(context.Context, models.ID) (map[string]int, error) {
	return map[string]int{"projects": 2}, nil
}
func (insightStoreFake) Audit(context.Context, models.ID, int, int) ([]models.AuditEvent, int, error) {
	return []models.AuditEvent{{ID: "event-id"}}, 1, nil
}

type membershipsFake struct{ role models.Role }

func (m membershipsFake) Role(context.Context, models.ID, models.ID) (models.Role, error) {
	if m.role == "" {
		return "", domain.ErrForbidden
	}
	return m.role, nil
}
func (membershipsFake) Get(context.Context, models.ID) (models.Project, error) {
	return models.Project{}, nil
}
func (membershipsFake) List(context.Context, database.ProjectFilter) ([]models.Project, int, error) {
	return nil, 0, nil
}
func (membershipsFake) Create(context.Context, models.Project) error { return nil }

func TestSummaryAndAudit(t *testing.T) {
	t.Parallel()
	service := NewService(insightStoreFake{}, membershipsFake{role: models.RoleViewer})
	summary, err := service.Summary(context.Background(), "user-id", "org-id")
	if err != nil || summary["projects"] != 2 {
		t.Fatalf("summary=%v err=%v", summary, err)
	}
	page, err := service.Audit(context.Background(), "user-id", "org-id", 0, 200)
	if err != nil || len(page.Items) != 1 || page.Page != 1 || page.PageSize != 100 || page.TotalPages != 1 {
		t.Fatalf("page=%+v err=%v", page, err)
	}
}

func TestInsightsRequireMembership(t *testing.T) {
	t.Parallel()
	service := NewService(insightStoreFake{}, membershipsFake{})
	if _, err := service.Summary(context.Background(), "user-id", "org-id"); err != domain.ErrForbidden {
		t.Fatalf("summary error=%v", err)
	}
	if _, err := service.Audit(context.Background(), "user-id", "org-id", 1, 20); err != domain.ErrForbidden {
		t.Fatalf("audit error=%v", err)
	}
}
