package repositories

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/forgeflow/forgeflow/internal/authorization"
	"github.com/forgeflow/forgeflow/internal/database"
	"github.com/forgeflow/forgeflow/internal/domain"
	"github.com/forgeflow/forgeflow/pkg/models"
	"github.com/forgeflow/forgeflow/pkg/validation"
	"github.com/google/uuid"
)

type Store interface{Create(context.Context,models.Repository)error;List(context.Context,models.ID)([]models.Repository,error)}
type Projects interface{Get(context.Context,models.ID)(models.Project,error);Role(context.Context,models.ID,models.ID)(models.Role,error);List(context.Context,database.ProjectFilter)([]models.Project,int,error);Create(context.Context,models.Project)error}
type Service struct{store Store;projects Projects;workspaceRoot string;now func()time.Time}
func NewService(store Store,projects Projects,workspaceRoot string)*Service{return &Service{store:store,projects:projects,workspaceRoot:workspaceRoot,now:time.Now}}
type CreateInput struct{Name string `json:"name"`;LocalPath string `json:"localPath"`}

func(s *Service)Create(ctx context.Context,actor models.User,projectID models.ID,input CreateInput)(models.Repository,error){
	if err:=s.authorize(ctx,actor.ID,projectID,authorization.Write);err!=nil{return models.Repository{},err};if !validation.Required(input.Name,100){return models.Repository{},domain.Invalid("name","must contain 1 to 100 characters")}
	path,err:=safePath(s.workspaceRoot,input.LocalPath);if err!=nil{return models.Repository{},domain.Invalid("localPath",err.Error())};info,err:=os.Stat(path);if err!=nil||!info.IsDir(){return models.Repository{},domain.Invalid("localPath","must be an existing directory beneath the workspace root")}
	branch,err:=gitOutput(ctx,path,"symbolic-ref","--quiet","--short","HEAD");if err!=nil{return models.Repository{},domain.Invalid("localPath","must be a Git repository with a checked-out branch")}
	repository:=models.Repository{ID:models.ID(uuid.NewString()),ProjectID:projectID,Name:strings.TrimSpace(input.Name),LocalPath:path,DefaultBranch:strings.TrimSpace(branch),CreatedAt:s.now().UTC()};if err:=s.store.Create(ctx,repository);err!=nil{return models.Repository{},err};return repository,nil
}
func(s *Service)List(ctx context.Context,actor models.User,projectID models.ID)([]models.Repository,error){if err:=s.authorize(ctx,actor.ID,projectID,authorization.Read);err!=nil{return nil,err};return s.store.List(ctx,projectID)}
func(s *Service)authorize(ctx context.Context,userID,projectID models.ID,permission authorization.Permission)error{project,err:=s.projects.Get(ctx,projectID);if err!=nil{return err};role,err:=s.projects.Role(ctx,project.OrganizationID,userID);if err!=nil{return err};return authorization.Require(role,permission)}

func safePath(root,input string)(string,error){if strings.TrimSpace(input)==""{return "",errors.New("is required")};root,err:=filepath.Abs(root);if err!=nil{return "",err};root,err=filepath.EvalSymlinks(root);if err!=nil{return "",fmt.Errorf("workspace root is unavailable")};candidate:=input;if !filepath.IsAbs(candidate){candidate=filepath.Join(root,candidate)};candidate,err=filepath.Abs(candidate);if err!=nil{return "",err};candidate,err=filepath.EvalSymlinks(candidate);if err!=nil{return "",fmt.Errorf("path is unavailable")};relative,err:=filepath.Rel(root,candidate);if err!=nil||relative==".."||strings.HasPrefix(relative,".."+string(filepath.Separator)){return "",errors.New("must resolve beneath the configured workspace root")};return candidate,nil}
func gitOutput(ctx context.Context,directory string,args ...string)(string,error){all:=append([]string{"-C",directory},args...);command:=exec.CommandContext(ctx,"git",all...);output,err:=command.Output();if err!=nil{return "",fmt.Errorf("git %s: %w",args[0],err)};return string(output),nil}

