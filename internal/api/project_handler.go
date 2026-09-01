package api

import (
	"net/http"
	"strconv"

	"github.com/forgeflow/forgeflow/internal/projects"
	"github.com/forgeflow/forgeflow/internal/tasks"
	"github.com/forgeflow/forgeflow/pkg/models"
	"github.com/go-chi/chi/v5"
)

type projectHandler struct {
	projects *projects.Service
	tasks    *tasks.Service
}

func (h projectHandler) create(w http.ResponseWriter, r *http.Request) {
	var input projects.CreateInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, r, err)
		return
	}
	project, err := h.projects.Create(r.Context(), currentUser(r.Context()), input)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, project)
}
func (h projectHandler) get(w http.ResponseWriter, r *http.Request) {
	project, err := h.projects.Get(r.Context(), currentUser(r.Context()), models.ID(chi.URLParam(r, "projectID")))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, project)
}
func (h projectHandler) list(w http.ResponseWriter, r *http.Request) {
	page, pageSize := pagination(r)
	result, err := h.projects.List(r.Context(), currentUser(r.Context()), projects.ListInput{OrganizationID: models.ID(r.URL.Query().Get("organizationId")), Search: r.URL.Query().Get("search"), Sort: r.URL.Query().Get("sort"), Desc: r.URL.Query().Get("order") == "desc", Page: page, PageSize: pageSize})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
func (h projectHandler) createTask(w http.ResponseWriter, r *http.Request) {
	var input tasks.CreateInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, r, err)
		return
	}
	task, err := h.tasks.Create(r.Context(), currentUser(r.Context()), models.ID(chi.URLParam(r, "projectID")), input)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, task)
}
func (h projectHandler) listTasks(w http.ResponseWriter, r *http.Request) {
	page, pageSize := pagination(r)
	result, err := h.tasks.List(r.Context(), currentUser(r.Context()), models.ID(chi.URLParam(r, "projectID")), models.TaskStatus(r.URL.Query().Get("status")), page, pageSize)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func pagination(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return page, size
}
