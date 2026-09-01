package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type contextKey string

const requestIDKey contextKey = "request-id"

func New(level string) *slog.Logger {
	var parsed slog.Level
	if err := parsed.UnmarshalText([]byte(level)); err != nil {
		parsed = slog.LevelInfo
	}
	return slog.New(&redactingHandler{Handler: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parsed})})
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func RequestID(ctx context.Context) string {
	value, _ := ctx.Value(requestIDKey).(string)
	return value
}

type redactingHandler struct{ slog.Handler }

func (h *redactingHandler) Handle(ctx context.Context, record slog.Record) error {
	clean := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	if id := RequestID(ctx); id != "" {
		clean.AddAttrs(slog.String("request_id", id))
	}
	record.Attrs(func(attr slog.Attr) bool {
		clean.AddAttrs(redact(attr))
		return true
	})
	return h.Handler.Handle(ctx, clean)
}

func redact(attr slog.Attr) slog.Attr {
	key := strings.ToLower(attr.Key)
	for _, sensitive := range []string{"password", "secret", "token", "authorization"} {
		if strings.Contains(key, sensitive) {
			return slog.String(attr.Key, "[REDACTED]")
		}
	}
	if attr.Value.Kind() == slog.KindGroup {
		items := attr.Value.Group()
		for i := range items {
			items[i] = redact(items[i])
		}
		return slog.Group(attr.Key, items...)
	}
	return attr
}

