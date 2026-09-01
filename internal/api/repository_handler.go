package api

import (
	"net/http"

	"github.com/forgeflow/forgeflow/internal/repositories"
	"github.com/forgeflow/forgeflow/pkg/models"
	"github.com/go-chi/chi/v5"
)

type repositoryHandler struct{ service *repositories.Service }

func (h repositoryHandler) create(w http.ResponseWriter, r *http.Request) {
	var input repositories.CreateInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, r, err)
		return
	}
	repository, err := h.service.Create(r.Context(), currentUser(r.Context()), models.ID(chi.URLParam(r, "projectID")), input)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, repository)
}
func (h repositoryHandler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context(), currentUser(r.Context()), models.ID(chi.URLParam(r, "projectID")))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
