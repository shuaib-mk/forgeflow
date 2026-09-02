package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/forgeflow/forgeflow/internal/domain"
	"github.com/forgeflow/forgeflow/pkg/models"
	"github.com/jackc/pgx/v5"
)

type WorkflowRepository struct{ db *DB }

func NewWorkflowRepository(db *DB) *WorkflowRepository { return &WorkflowRepository{db: db} }

func (r *WorkflowRepository) Create(ctx context.Context, workflow models.Workflow) error {
	steps, err := json.Marshal(workflow.Steps)
	if err != nil {
		return fmt.Errorf("encode workflow: %w", err)
	}
	_, err = r.db.Pool.Exec(ctx, `INSERT INTO workflows (id,project_id,name,version,steps,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$6)`, workflow.ID, workflow.ProjectID, workflow.Name, workflow.Version, steps, workflow.CreatedAt)
	if err != nil {
		return fmt.Errorf("create workflow: %w", err)
	}
	return nil
}

func (r *WorkflowRepository) Get(ctx context.Context, id models.ID) (models.Workflow, error) {
	var workflow models.Workflow
	var steps []byte
	err := r.db.Pool.QueryRow(ctx, `SELECT id,project_id,name,version,steps,created_at,updated_at FROM workflows WHERE id=$1`, id).Scan(&workflow.ID, &workflow.ProjectID, &workflow.Name, &workflow.Version, &steps, &workflow.CreatedAt, &workflow.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Workflow{}, domain.ErrNotFound
	}
	if err != nil {
		return models.Workflow{}, fmt.Errorf("get workflow: %w", err)
	}
	if err := json.Unmarshal(steps, &workflow.Steps); err != nil {
		return models.Workflow{}, fmt.Errorf("decode workflow: %w", err)
	}
	return workflow, nil
}

func (r *WorkflowRepository) ExecutionDirectory(ctx context.Context, projectID models.ID) (string, error) {
	var path string
	err := r.db.Pool.QueryRow(ctx, `SELECT local_path FROM repositories WHERE project_id=$1 ORDER BY created_at,id LIMIT 1`, projectID).Scan(&path)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get workflow execution directory: %w", err)
	}
	return path, nil
}

func (r *WorkflowRepository) ListWorkflows(ctx context.Context, projectID models.ID) ([]models.Workflow, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT id,project_id,name,version,steps,created_at,updated_at FROM workflows WHERE project_id=$1 ORDER BY updated_at DESC,id`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list workflows: %w", err)
	}
	defer rows.Close()
	items := []models.Workflow{}
	for rows.Next() {
		var workflow models.Workflow
		var steps []byte
		if err := rows.Scan(&workflow.ID, &workflow.ProjectID, &workflow.Name, &workflow.Version, &steps, &workflow.CreatedAt, &workflow.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan workflow: %w", err)
		}
		if err := json.Unmarshal(steps, &workflow.Steps); err != nil {
			return nil, fmt.Errorf("decode workflow: %w", err)
		}
		items = append(items, workflow)
	}
	return items, rows.Err()
}

func (r *WorkflowRepository) CreateRun(ctx context.Context, run models.WorkflowRun) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO workflow_runs (id,workflow_id,project_id,status,triggered_by,attempt,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`, run.ID, run.WorkflowID, run.ProjectID, run.Status, run.TriggeredBy, run.Attempt, run.CreatedAt)
	if err != nil {
		return fmt.Errorf("create workflow run: %w", err)
	}
	return nil
}

func (r *WorkflowRepository) GetRun(ctx context.Context, id models.ID) (models.WorkflowRun, error) {
	var run models.WorkflowRun
	err := r.db.Pool.QueryRow(ctx, `SELECT id,workflow_id,project_id,status,triggered_by,attempt,started_at,finished_at,error,created_at FROM workflow_runs WHERE id=$1`, id).Scan(&run.ID, &run.WorkflowID, &run.ProjectID, &run.Status, &run.TriggeredBy, &run.Attempt, &run.StartedAt, &run.FinishedAt, &run.Error, &run.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.WorkflowRun{}, domain.ErrNotFound
	}
	if err != nil {
		return models.WorkflowRun{}, fmt.Errorf("get workflow run: %w", err)
	}
	return run, nil
}

func (r *WorkflowRepository) ListRuns(ctx context.Context, organizationID, projectID models.ID, limit, offset int) ([]models.WorkflowRun, int, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT wr.id,wr.workflow_id,wr.project_id,wr.status,wr.triggered_by,wr.attempt,wr.started_at,wr.finished_at,wr.error,wr.created_at,count(*) OVER() FROM workflow_runs wr JOIN projects p ON p.id=wr.project_id WHERE p.organization_id=$1 AND ($2='' OR wr.project_id::text=$2) ORDER BY wr.created_at DESC,wr.id LIMIT $3 OFFSET $4`, organizationID, projectID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list workflow runs: %w", err)
	}
	defer rows.Close()
	items := []models.WorkflowRun{}
	total := 0
	for rows.Next() {
		var run models.WorkflowRun
		if err := rows.Scan(&run.ID, &run.WorkflowID, &run.ProjectID, &run.Status, &run.TriggeredBy, &run.Attempt, &run.StartedAt, &run.FinishedAt, &run.Error, &run.CreatedAt, &total); err != nil {
			return nil, 0, fmt.Errorf("scan workflow run: %w", err)
		}
		items = append(items, run)
	}
	return items, total, rows.Err()
}

func (r *WorkflowRepository) ClaimRun(ctx context.Context, id models.ID) (models.WorkflowRun, bool, error) {
	var run models.WorkflowRun
	now := time.Now().UTC()
	err := r.db.Pool.QueryRow(ctx, `UPDATE workflow_runs SET status='running',started_at=$2,attempt=attempt+1 WHERE id=$1 AND status='queued' RETURNING id,workflow_id,project_id,status,triggered_by,attempt,started_at,finished_at,error,created_at`, id, now).Scan(&run.ID, &run.WorkflowID, &run.ProjectID, &run.Status, &run.TriggeredBy, &run.Attempt, &run.StartedAt, &run.FinishedAt, &run.Error, &run.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.WorkflowRun{}, false, nil
	}
	if err != nil {
		return models.WorkflowRun{}, false, fmt.Errorf("claim workflow run: %w", err)
	}
	return run, true, nil
}

func (r *WorkflowRepository) CompleteRun(ctx context.Context, id models.ID, status models.RunStatus, message string) error {
	result, err := r.db.Pool.Exec(ctx, `UPDATE workflow_runs SET status=$2,error=$3,finished_at=now() WHERE id=$1 AND status='running'`, id, status, message)
	if err != nil {
		return fmt.Errorf("complete workflow run: %w", err)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrConflict
	}
	return nil
}

func (r *WorkflowRepository) AppendLog(ctx context.Context, runID models.ID, stepID string, sequence int, content string) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO run_logs (run_id,step_id,sequence,content) VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`, runID, stepID, sequence, content)
	if err != nil {
		return fmt.Errorf("append run log: %w", err)
	}
	return nil
}

func (r *WorkflowRepository) Logs(ctx context.Context, runID models.ID, after int) ([]string, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT content FROM run_logs WHERE run_id=$1 AND sequence>$2 ORDER BY sequence LIMIT 1000`, runID, after)
	if err != nil {
		return nil, fmt.Errorf("list run logs: %w", err)
	}
	defer rows.Close()
	logs := []string{}
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return nil, err
		}
		logs = append(logs, line)
	}
	return logs, rows.Err()
}
