package database

import (
	"context"
	"fmt"

	"github.com/forgeflow/forgeflow/pkg/models"
)

type RepositoryRepository struct{db *DB}
func NewRepositoryRepository(db *DB)*RepositoryRepository{return &RepositoryRepository{db:db}}
func(r *RepositoryRepository)Create(ctx context.Context,repository models.Repository)error{_,err:=r.db.Pool.Exec(ctx,`INSERT INTO repositories(id,project_id,name,local_path,default_branch,created_at) VALUES($1,$2,$3,$4,$5,$6)`,repository.ID,repository.ProjectID,repository.Name,repository.LocalPath,repository.DefaultBranch,repository.CreatedAt);if err!=nil{return fmt.Errorf("create repository: %w",err)};return nil}
func(r *RepositoryRepository)List(ctx context.Context,projectID models.ID)([]models.Repository,error){rows,err:=r.db.Pool.Query(ctx,`SELECT id,project_id,name,local_path,default_branch,created_at FROM repositories WHERE project_id=$1 ORDER BY name`,projectID);if err!=nil{return nil,fmt.Errorf("list repositories: %w",err)};defer rows.Close();items:=[]models.Repository{};for rows.Next(){var item models.Repository;if err:=rows.Scan(&item.ID,&item.ProjectID,&item.Name,&item.LocalPath,&item.DefaultBranch,&item.CreatedAt);err!=nil{return nil,err};items=append(items,item)};return items,rows.Err()}

