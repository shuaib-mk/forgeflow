package api

import (
	"net/http"
	"strconv"

	"github.com/forgeflow/forgeflow/internal/workflows"
	"github.com/forgeflow/forgeflow/pkg/models"
	"github.com/go-chi/chi/v5"
)

type workflowHandler struct{ service *workflows.Service }

func (h workflowHandler) create(w http.ResponseWriter, r *http.Request) {
	var definition workflows.Definition
	if err := decodeJSON(w, r, &definition); err != nil {
		writeError(w, r, err)
		return
	}
	workflow, err := h.service.Create(r.Context(), currentUser(r.Context()), models.ID(chi.URLParam(r, "projectID")), definition)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, workflow)
}
func (h workflowHandler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context(), currentUser(r.Context()), models.ID(chi.URLParam(r, "projectID")))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
func (h workflowHandler) run(w http.ResponseWriter, r *http.Request) {
	run, err := h.service.Run(r.Context(), currentUser(r.Context()), models.ID(chi.URLParam(r, "workflowID")))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusAccepted, run)
}
func (h workflowHandler) getRun(w http.ResponseWriter, r *http.Request) {
	run, err := h.service.GetRun(r.Context(), currentUser(r.Context()), models.ID(chi.URLParam(r, "runID")))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}
func (h workflowHandler) listRuns(w http.ResponseWriter, r *http.Request) {
	page, size := pagination(r)
	runs, err := h.service.ListRuns(r.Context(), currentUser(r.Context()), models.ID(r.URL.Query().Get("organizationId")), models.ID(r.URL.Query().Get("projectId")), page, size)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, runs)
}
func (h workflowHandler) logs(w http.ResponseWriter, r *http.Request) {
	after, _ := strconv.Atoi(r.URL.Query().Get("after"))
	logs, err := h.service.Logs(r.Context(), currentUser(r.Context()), models.ID(chi.URLParam(r, "runID")), after)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": logs, "next": after + len(logs)})
}
