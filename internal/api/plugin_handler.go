package api

import (
	"context"
	"net/http"

	"github.com/forgeflow/forgeflow/pkg/models"
)

type PluginStore interface{List(context.Context)([]models.Plugin,error)}
type pluginHandler struct{store PluginStore}
func(h pluginHandler)list(w http.ResponseWriter,r *http.Request){items,err:=h.store.List(r.Context());if err!=nil{writeError(w,r,err);return};writeJSON(w,http.StatusOK,map[string]any{"items":items})}

