package database

import (
	"context"
	"fmt"

	"github.com/forgeflow/forgeflow/pkg/models"
)

type PluginRepository struct{ db *DB }

func NewPluginRepository(db *DB) *PluginRepository { return &PluginRepository{db: db} }
func (r *PluginRepository) List(ctx context.Context) ([]models.Plugin, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT id,name,version,description,enabled,created_at FROM plugins ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list plugins: %w", err)
	}
	defer rows.Close()
	items := []models.Plugin{}
	for rows.Next() {
		var item models.Plugin
		if err := rows.Scan(&item.ID, &item.Name, &item.Version, &item.Description, &item.Enabled, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
