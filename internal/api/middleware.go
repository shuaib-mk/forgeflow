package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/forgeflow/forgeflow/internal/auth"
	"github.com/forgeflow/forgeflow/internal/domain"
	"github.com/forgeflow/forgeflow/pkg/logger"
	"github.com/forgeflow/forgeflow/pkg/models"
	"github.com/google/uuid"
)

type userKeyType string
type loggerKeyType string

const userKey userKeyType = "authenticated-user"
const serverLoggerKey loggerKeyType = "server-logger"

func withLogger(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), serverLoggerKey, log)))
	})
}

func requestLogger(ctx context.Context) *slog.Logger {
	if log, ok := ctx.Value(serverLoggerKey).(*slog.Logger); ok && log != nil {
		return log
	}
	return slog.Default()
}

func requestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if len(id) < 8 || len(id) > 128 {
			id = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(logger.WithRequestID(r.Context(), id)))
	})
}

func recoverer(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if value := recover(); value != nil {
				log.ErrorContext(r.Context(), "panic recovered", "panic", value, "stack", string(debug.Stack()))
				writeError(w, r, errors.New("panic recovered"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func accessLog(log *slog.Logger, metrics *Metrics, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		duration := time.Since(start)
		metrics.Observe(r.Method, r.URL.Path, wrapped.status, duration)
		log.InfoContext(r.Context(), "http request", "method", r.Method, "path", r.URL.Path, "status", wrapped.status, "duration_ms", duration.Milliseconds())
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func authenticate(service *auth.Service, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			writeError(w, r, domain.ErrUnauthorized)
			return
		}
		user, err := service.Authenticate(r.Context(), strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
		if err != nil {
			writeError(w, r, err)
			return
		}
		ctx := context.WithValue(r.Context(), userKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func currentUser(ctx context.Context) models.User {
	user, _ := ctx.Value(userKey).(models.User)
	return user
}
