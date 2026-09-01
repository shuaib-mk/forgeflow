package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelpListsCoreCommands(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	command := NewRoot(&output, &output)
	command.SetArgs([]string{"--help"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"project", "task", "workflow", "doctor", "config"} {
		if !strings.Contains(output.String(), name) {
			t.Errorf("help missing %q", name)
		}
	}
}

func TestWorkflowValidate(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	path := filepath.Join(directory, "workflow.yaml")
	if err := os.WriteFile(path, []byte("name: ci\nsteps:\n  - id: test\n    command: go\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	command := NewRoot(&output, &output)
	command.SetArgs([]string{"workflow", "validate", path})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "is valid") {
		t.Fatalf("output=%s", output.String())
	}
}
