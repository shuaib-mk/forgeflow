package example

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/forgeflow/forgeflow/internal/plugins"
)

type Summary struct{}

func (Summary) Manifest() plugins.Manifest {
	return plugins.Manifest{Name: "run-summary", Version: "v1.0.0", Description: "Writes structured workflow summaries for downstream tooling"}
}

func (Summary) Register(registrar *plugins.Registrar) error {
	return registrar.WorkflowStep("summary.write", func(_ context.Context, input plugins.StepInput) (plugins.StepOutput, error) {
		if len(input.Args) != 2 { return plugins.StepOutput{}, fmt.Errorf("summary.write expects a file name and message") }
		path := filepath.Join(input.WorkingDirectory, filepath.Base(input.Args[0]))
		content, err := json.MarshalIndent(map[string]string{"summary": input.Args[1]}, "", "  ")
		if err != nil { return plugins.StepOutput{}, err }
		if err := os.WriteFile(path, append(content, '\n'), 0o600); err != nil { return plugins.StepOutput{}, fmt.Errorf("write summary: %w", err) }
		return plugins.StepOutput{Logs: "summary written", Metadata: map[string]string{"path": path}}, nil
	})
}

func (Summary) Close(context.Context) error { return nil }

