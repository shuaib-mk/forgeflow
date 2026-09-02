package api

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type healthy struct{}

func (healthy) Ping(context.Context) error { return nil }

func TestHealthAndReadiness(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewServer(Dependencies{Database: healthy{}, Queue: healthy{}, Logger: log, AllowedOrigins: []string{"http://localhost"}})
	for _, path := range []string{"/health", "/ready"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		server.Handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s status=%d body=%s", path, response.Code, response.Body.String())
		}
		if response.Header().Get("X-Request-ID") == "" {
			t.Fatal("request id missing")
		}
	}
}

func TestDecodeJSONRejectsUnknownAndMultipleValues(t *testing.T) {
	t.Parallel()
	for _, body := range []string{`{"known":true,"unknown":1}`, `{"known":true}{"known":false}`} {
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		response := httptest.NewRecorder()
		var input struct {
			Known bool `json:"known"`
		}
		if err := decodeJSON(response, request, &input); err == nil {
			t.Fatalf("body %q expected error", body)
		}
	}
}

func TestWriteErrorLogsInternalFailureWithRequestContext(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&output, nil))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	request = request.WithContext(context.WithValue(request.Context(), serverLoggerKey, log))
	response := httptest.NewRecorder()
	writeError(response, request, errors.New("database unavailable"))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d", response.Code)
	}
	for _, expected := range []string{"request failed", "database unavailable", "/api/v1/projects"} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("log=%q missing %q", output.String(), expected)
		}
	}
}
