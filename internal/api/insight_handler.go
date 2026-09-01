package api

import (
	"net/http"

	"github.com/forgeflow/forgeflow/internal/analytics"
	"github.com/forgeflow/forgeflow/pkg/models"
)

type insightHandler struct{service *analytics.Service}
func(h insightHandler)analytics(w http.ResponseWriter,r *http.Request){summary,err:=h.service.Summary(r.Context(),currentUser(r.Context()).ID,models.ID(r.URL.Query().Get("organizationId")));if err!=nil{writeError(w,r,err);return};writeJSON(w,http.StatusOK,summary)}
func(h insightHandler)audit(w http.ResponseWriter,r *http.Request){page,size:=pagination(r);events,err:=h.service.Audit(r.Context(),currentUser(r.Context()).ID,models.ID(r.URL.Query().Get("organizationId")),page,size);if err!=nil{writeError(w,r,err);return};writeJSON(w,http.StatusOK,events)}

