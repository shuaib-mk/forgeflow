package analytics

import (
	"context"
	"math"

	"github.com/forgeflow/forgeflow/internal/authorization"
	"github.com/forgeflow/forgeflow/internal/database"
	"github.com/forgeflow/forgeflow/pkg/models"
)

type Store interface{Analytics(context.Context,models.ID)(map[string]int,error);Audit(context.Context,models.ID,int,int)([]models.AuditEvent,int,error)}
type Memberships interface{Role(context.Context,models.ID,models.ID)(models.Role,error);Get(context.Context,models.ID)(models.Project,error);List(context.Context,database.ProjectFilter)([]models.Project,int,error);Create(context.Context,models.Project)error}
type Service struct{store Store;memberships Memberships}
func NewService(store Store,memberships Memberships)*Service{return &Service{store:store,memberships:memberships}}
func(s *Service)Summary(ctx context.Context,userID,organizationID models.ID)(map[string]int,error){role,err:=s.memberships.Role(ctx,organizationID,userID);if err!=nil{return nil,err};if err:=authorization.Require(role,authorization.Read);err!=nil{return nil,err};return s.store.Analytics(ctx,organizationID)}
func(s *Service)Audit(ctx context.Context,userID,organizationID models.ID,page,pageSize int)(models.Page[models.AuditEvent],error){role,err:=s.memberships.Role(ctx,organizationID,userID);if err!=nil{return models.Page[models.AuditEvent]{},err};if err:=authorization.Require(role,authorization.Read);err!=nil{return models.Page[models.AuditEvent]{},err};if page<1{page=1};if pageSize<1{pageSize=20};if pageSize>100{pageSize=100};items,total,err:=s.store.Audit(ctx,organizationID,pageSize,(page-1)*pageSize);if err!=nil{return models.Page[models.AuditEvent]{},err};return models.Page[models.AuditEvent]{Items:items,Page:page,PageSize:pageSize,TotalItems:total,TotalPages:int(math.Ceil(float64(total)/float64(pageSize)))},nil}

