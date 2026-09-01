package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Environment       string
	HTTPAddr          string
	DatabaseURL       string
	RedisURL          string
	SessionSecret     string
	WorkspaceRoot     string
	LogLevel          string
	WorkerConcurrency int
	AllowedOrigins    []string
}

func Load() (Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return Config{}, fmt.Errorf("working directory: %w", err)
	}
	concurrency, err := strconv.Atoi(get("FORGEFLOW_WORKER_CONCURRENCY", "2"))
	if err != nil {
		return Config{}, fmt.Errorf("FORGEFLOW_WORKER_CONCURRENCY: %w", err)
	}
	cfg := Config{
		Environment:       get("FORGEFLOW_ENV", "development"),
		HTTPAddr:          get("FORGEFLOW_HTTP_ADDR", ":8080"),
		DatabaseURL:       get("FORGEFLOW_DATABASE_URL", "postgres://forgeflow:forgeflow@localhost:5432/forgeflow?sslmode=disable"),
		RedisURL:          get("FORGEFLOW_REDIS_URL", "redis://localhost:6379/0"),
		SessionSecret:     os.Getenv("FORGEFLOW_SESSION_SECRET"),
		WorkspaceRoot:     get("FORGEFLOW_WORKSPACE_ROOT", filepath.Join(cwd, "data", "workspaces")),
		LogLevel:          get("FORGEFLOW_LOG_LEVEL", "info"),
		WorkerConcurrency: concurrency,
		AllowedOrigins:    splitCSV(get("FORGEFLOW_ALLOWED_ORIGINS", "http://localhost:5173")),
	}
	return cfg, cfg.Validate()
}

func (c Config) Validate() error {
	var errs []error
	if _, err := url.ParseRequestURI(c.DatabaseURL); err != nil {
		errs = append(errs, errors.New("FORGEFLOW_DATABASE_URL must be a valid URL"))
	}
	if _, err := url.ParseRequestURI(c.RedisURL); err != nil {
		errs = append(errs, errors.New("FORGEFLOW_REDIS_URL must be a valid URL"))
	}
	if len(c.SessionSecret) < 32 {
		errs = append(errs, errors.New("FORGEFLOW_SESSION_SECRET must contain at least 32 characters"))
	}
	if c.WorkerConcurrency < 1 || c.WorkerConcurrency > 32 {
		errs = append(errs, errors.New("FORGEFLOW_WORKER_CONCURRENCY must be between 1 and 32"))
	}
	if len(c.AllowedOrigins) == 0 {
		errs = append(errs, errors.New("FORGEFLOW_ALLOWED_ORIGINS must not be empty"))
	}
	return errors.Join(errs...)
}

func get(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func splitCSV(value string) []string {
	var values []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			values = append(values, item)
		}
	}
	return values
}
