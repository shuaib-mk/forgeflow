package config

import "testing"

func TestValidateRejectsUnsafeConfiguration(t *testing.T) {
	t.Parallel()
	cfg := Config{DatabaseURL: "://", RedisURL: "://", SessionSecret: "short", WorkerConcurrency: 0}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() expected an error")
	}
}

func TestValidateAcceptsCompleteConfiguration(t *testing.T) {
	t.Parallel()
	cfg := Config{
		DatabaseURL:       "postgres://localhost/forgeflow",
		RedisURL:          "redis://localhost:6379/0",
		SessionSecret:     "01234567890123456789012345678901",
		WorkerConcurrency: 2,
		AllowedOrigins:    []string{"http://localhost:5173"},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}
}
