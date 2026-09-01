package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/forgeflow/forgeflow/internal/domain"
	"github.com/forgeflow/forgeflow/pkg/models"
	"github.com/jackc/pgx/v5"
)

type ProjectFilter struct {
	OrganizationID models.ID
	Search         string
	Sort           string
	Desc           bool
	Limit, Offset  int
}
type ProjectRepository struct{ db *DB }

func NewProjectRepository(db *DB) *ProjectRepository { return &ProjectRepository{db: db} }

func (r *ProjectRepository) Create(ctx context.Context, project models.Project) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO projects (id,organization_id,name,slug,description,created_by,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$7)`, project.ID, project.OrganizationID, project.Name, project.Slug, project.Description, project.CreatedBy, project.CreatedAt)
	if err != nil {
		return fmt.Errorf("create project: %w", err)
	}
	return nil
}

func (r *ProjectRepository) Get(ctx context.Context, id models.ID) (models.Project, error) {
	var project models.Project
	err := r.db.Pool.QueryRow(ctx, `SELECT id,organization_id,name,slug,description,created_by,created_at,updated_at FROM projects WHERE id=$1`, id).Scan(&project.ID, &project.OrganizationID, &project.Name, &project.Slug, &project.Description, &project.CreatedBy, &project.CreatedAt, &project.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Project{}, domain.ErrNotFound
	}
	if err != nil {
		return models.Project{}, fmt.Errorf("get project: %w", err)
	}
	return project, nil
}

func (r *ProjectRepository) List(ctx context.Context, filter ProjectFilter) ([]models.Project, int, error) {
	order := "created_at"
	if filter.Sort == "name" {
		order = "name"
	}
	direction := "ASC"
	if filter.Desc {
		direction = "DESC"
	}
	query := fmt.Sprintf(`SELECT id,organization_id,name,slug,description,created_by,created_at,updated_at,count(*) OVER() FROM projects WHERE organization_id=$1 AND ($2='' OR name ILIKE '%%'||$2||'%%') ORDER BY %s %s, id ASC LIMIT $3 OFFSET $4`, order, direction)
	rows, err := r.db.Pool.Query(ctx, query, filter.OrganizationID, filter.Search, filter.Limit, filter.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()
	items := []models.Project{}
	total := 0
	for rows.Next() {
		var project models.Project
		if err := rows.Scan(&project.ID, &project.OrganizationID, &project.Name, &project.Slug, &project.Description, &project.CreatedBy, &project.CreatedAt, &project.UpdatedAt, &total); err != nil {
			return nil, 0, fmt.Errorf("scan project: %w", err)
		}
		items = append(items, project)
	}
	return items, total, rows.Err()
}

func (r *ProjectRepository) Role(ctx context.Context, organizationID, userID models.ID) (models.Role, error) {
	var role models.Role
	err := r.db.Pool.QueryRow(ctx, `SELECT role FROM memberships WHERE organization_id=$1 AND user_id=$2`, organizationID, userID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrForbidden
	}
	if err != nil {
		return "", fmt.Errorf("get membership: %w", err)
	}
	return role, nil
}
