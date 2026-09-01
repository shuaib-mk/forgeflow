package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/forgeflow/forgeflow/internal/domain"
	"github.com/forgeflow/forgeflow/pkg/models"
	"github.com/jackc/pgx/v5"
)

type TaskRepository struct{ db *DB }
func NewTaskRepository(db *DB) *TaskRepository { return &TaskRepository{db: db} }

func (r *TaskRepository) Create(ctx context.Context, task models.Task) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO tasks (id,project_id,title,description,status,assignee_id,created_by,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$8)`, task.ID, task.ProjectID, task.Title, task.Description, task.Status, task.AssigneeID, task.CreatedBy, task.CreatedAt)
	if err != nil { return fmt.Errorf("create task: %w", err) }
	return nil
}

func (r *TaskRepository) List(ctx context.Context, projectID models.ID, status models.TaskStatus, limit, offset int) ([]models.Task, int, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT id,project_id,title,description,status,assignee_id,created_by,created_at,updated_at,count(*) OVER() FROM tasks WHERE project_id=$1 AND ($2='' OR status=$2) ORDER BY created_at DESC LIMIT $3 OFFSET $4`, projectID, status, limit, offset)
	if err != nil { return nil, 0, fmt.Errorf("list tasks: %w", err) }
	defer rows.Close()
	items := []models.Task{}; total := 0
	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.ID,&task.ProjectID,&task.Title,&task.Description,&task.Status,&task.AssigneeID,&task.CreatedBy,&task.CreatedAt,&task.UpdatedAt,&total); err != nil { return nil,0,fmt.Errorf("scan task: %w",err) }
		items = append(items, task)
	}
	return items,total,rows.Err()
}

func (r *TaskRepository) UpdateStatus(ctx context.Context, id models.ID, status models.TaskStatus) (models.Task, error) {
	var task models.Task
	err := r.db.Pool.QueryRow(ctx, `UPDATE tasks SET status=$2,updated_at=now() WHERE id=$1 RETURNING id,project_id,title,description,status,assignee_id,created_by,created_at,updated_at`, id, status).Scan(&task.ID,&task.ProjectID,&task.Title,&task.Description,&task.Status,&task.AssigneeID,&task.CreatedBy,&task.CreatedAt,&task.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) { return models.Task{}, domain.ErrNotFound }
	if err != nil { return models.Task{}, fmt.Errorf("update task: %w",err) }
	return task,nil
}

