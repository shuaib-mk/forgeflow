package audit

import (
	"context"
	"fmt"

	"github.com/forgeflow/forgeflow/internal/events"
	"github.com/forgeflow/forgeflow/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Subscribe(bus *events.Bus,pool *pgxpool.Pool)func(){
	return bus.Subscribe("project.created",func(ctx context.Context,event events.Event)error{
		created,ok:=event.(events.ProjectCreated);if !ok{return fmt.Errorf("unexpected project.created payload %T",event)}
		_,err:=pool.Exec(ctx,`INSERT INTO audit_events(id,organization_id,actor_id,action,resource_type,resource_id,request_id) SELECT $1,organization_id,$2,'project.created','project',$3,$4 FROM projects WHERE id=$3`,uuid.NewString(),created.ActorID,created.ProjectID,logger.RequestID(ctx));return err
	})
}

