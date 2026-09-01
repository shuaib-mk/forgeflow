package database

import (
	"context"
	"fmt"

	"github.com/forgeflow/forgeflow/pkg/models"
)

type InsightsRepository struct{ db *DB }

func NewInsightsRepository(db *DB) *InsightsRepository { return &InsightsRepository{db: db} }
func (r *InsightsRepository) Analytics(ctx context.Context, organizationID models.ID) (map[string]int, error) {
	var projects, openTasks, runningWorkflows, failedRuns int
	err := r.db.Pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM projects WHERE organization_id=$1),(SELECT count(*) FROM tasks t JOIN projects p ON p.id=t.project_id WHERE p.organization_id=$1 AND t.status<>'done'),(SELECT count(*) FROM workflow_runs wr JOIN projects p ON p.id=wr.project_id WHERE p.organization_id=$1 AND wr.status='running'),(SELECT count(*) FROM workflow_runs wr JOIN projects p ON p.id=wr.project_id WHERE p.organization_id=$1 AND wr.status IN ('failed','timed_out'))`, organizationID).Scan(&projects, &openTasks, &runningWorkflows, &failedRuns)
	if err != nil {
		return nil, fmt.Errorf("load analytics: %w", err)
	}
	return map[string]int{"projects": projects, "openTasks": openTasks, "runningWorkflows": runningWorkflows, "failedRuns": failedRuns}, nil
}
func (r *InsightsRepository) Audit(ctx context.Context, organizationID models.ID, limit, offset int) ([]models.AuditEvent, int, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT id,organization_id,actor_id,action,resource_type,resource_id,metadata,request_id,created_at,count(*) OVER() FROM audit_events WHERE organization_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, organizationID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("load audit events: %w", err)
	}
	defer rows.Close()
	items := []models.AuditEvent{}
	total := 0
	for rows.Next() {
		var item models.AuditEvent
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.ActorID, &item.Action, &item.ResourceType, &item.ResourceID, &item.Metadata, &item.RequestID, &item.CreatedAt, &total); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}
