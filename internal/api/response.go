package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/forgeflow/forgeflow/internal/domain"
	"github.com/forgeflow/forgeflow/pkg/logger"
)

type errorResponse struct {
	Error apiError `json:"error"`
}
type apiError struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	RequestID string            `json:"requestId"`
	Fields    map[string]string `json:"fields,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	code := "internal_error"
	message := "An unexpected error occurred."
	fields := map[string]string(nil)
	var validation *domain.ValidationError
	switch {
	case errors.As(err, &validation):
		status = http.StatusUnprocessableEntity
		code = "validation_error"
		message = "The request contains invalid fields."
		fields = validation.Fields
	case errors.Is(err, domain.ErrUnauthorized):
		status = http.StatusUnauthorized
		code = "unauthorized"
		message = "Authentication is required."
	case errors.Is(err, domain.ErrForbidden):
		status = http.StatusForbidden
		code = "forbidden"
		message = "You do not have permission to perform this action."
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
		code = "not_found"
		message = "The requested resource was not found."
	case errors.Is(err, domain.ErrConflict):
		status = http.StatusConflict
		code = "conflict"
		message = "The request conflicts with current resource state."
	}
	writeJSON(w, status, errorResponse{Error: apiError{Code: code, Message: message, RequestID: logger.RequestID(r.Context()), Fields: fields}})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return domain.Invalid("body", "must be valid JSON with known fields")
	}
	var extra any
	if decoder.Decode(&extra) == nil {
		return domain.Invalid("body", "must contain a single JSON object")
	}
	return nil
}
