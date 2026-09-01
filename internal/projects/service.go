package projects

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/forgeflow/forgeflow/internal/database"
	"github.com/forgeflow/forgeflow/internal/domain"
	"github.com/forgeflow/forgeflow/internal/events"
	"github.com/forgeflow/forgeflow/pkg/models"
	"github.com/forgeflow/forgeflow/pkg/validation"
	"github.com/google/uuid"
)

type Repository interface {
	Create(context.Context,models.Project) error
	Get(context.Context,models.ID)(models.Project,error)
	List(context.Context,database.ProjectFilter)([]models.Project,int,error)
	Role(context.Context,models.ID,models.ID)(models.Role,error)
}

type Service struct { repository Repository; events *events.Bus; now func()time.Time }
func NewService(repository Repository,bus *events.Bus)*Service{return &Service{repository:repository,events:bus,now:time.Now}}

type CreateInput struct{OrganizationID models.ID;Name,Slug,Description string}
type ListInput struct{OrganizationID models.ID;Search,Sort string;Desc bool;Page,PageSize int}

func (s *Service) Create(ctx context.Context,actor models.User,input CreateInput)(models.Project,error){
	role,err:=s.repository.Role(ctx,input.OrganizationID,actor.ID);if err!=nil{return models.Project{},err}
	if role!=models.RoleOwner&&role!=models.RoleAdmin&&role!=models.RoleDeveloper{return models.Project{},domain.ErrForbidden}
	if !validation.Required(input.Name,100){return models.Project{},domain.Invalid("name","must contain 1 to 100 characters")}
	if !validation.Slug(input.Slug){return models.Project{},domain.Invalid("slug","must be a lowercase URL-safe slug")}
	if len(input.Description)>2000{return models.Project{},domain.Invalid("description","must not exceed 2000 characters")}
	now:=s.now().UTC();project:=models.Project{ID:models.ID(uuid.NewString()),OrganizationID:input.OrganizationID,Name:strings.TrimSpace(input.Name),Slug:input.Slug,Description:strings.TrimSpace(input.Description),CreatedBy:actor.ID,CreatedAt:now,UpdatedAt:now}
	if err:=s.repository.Create(ctx,project);err!=nil{return models.Project{},err}
	if s.events!=nil{_ = s.events.Publish(ctx,events.ProjectCreated{ProjectID:project.ID,ActorID:actor.ID,At:now})}
	return project,nil
}

func (s *Service) Get(ctx context.Context,actor models.User,id models.ID)(models.Project,error){
	project,err:=s.repository.Get(ctx,id);if err!=nil{return models.Project{},err}
	if _,err:=s.repository.Role(ctx,project.OrganizationID,actor.ID);err!=nil{return models.Project{},err}
	return project,nil
}

func (s *Service) List(ctx context.Context,actor models.User,input ListInput)(models.Page[models.Project],error){
	if _,err:=s.repository.Role(ctx,input.OrganizationID,actor.ID);err!=nil{return models.Page[models.Project]{},err}
	if input.Page<1{input.Page=1};if input.PageSize<1{input.PageSize=20};if input.PageSize>100{input.PageSize=100}
	if input.Sort!="name"{input.Sort="createdAt"}
	items,total,err:=s.repository.List(ctx,database.ProjectFilter{OrganizationID:input.OrganizationID,Search:strings.TrimSpace(input.Search),Sort:map[string]string{"createdAt":"created_at","name":"name"}[input.Sort],Desc:input.Desc,Limit:input.PageSize,Offset:(input.Page-1)*input.PageSize})
	if err!=nil{return models.Page[models.Project]{},err}
	return models.Page[models.Project]{Items:items,Page:input.Page,PageSize:input.PageSize,TotalItems:total,TotalPages:int(math.Ceil(float64(total)/float64(input.PageSize)))},nil
}

