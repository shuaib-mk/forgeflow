package workflows

import (
	"strings"
	"testing"
	"time"
)

func TestParseValidWorkflow(t *testing.T) {
	t.Parallel()
	definition, err := Parse(strings.NewReader(`
name: test-and-build
steps:
  - id: lint
    name: Lint
    command: go
    args: [vet, ./...]
    timeout: 2m
  - id: test
    command: go
    args: [test, ./...]
    depends_on: [lint]
    retries: 1
`))
	if err != nil {
		t.Fatal(err)
	}
	if definition.Steps[0].Timeout != 2*time.Minute {
		t.Fatalf("timeout = %s", definition.Steps[0].Timeout)
	}
}

func TestValidateDetectsCycle(t *testing.T) {
	t.Parallel()
	_, err := Parse(strings.NewReader(`
name: cyclic
steps:
  - id: first
    command: echo
    depends_on: [second]
  - id: second
    command: echo
    depends_on: [first]
`))
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("expected cycle error, got %v", err)
	}
}

func TestParseRejectsUnknownFields(t *testing.T) {
	t.Parallel()
	_, err := Parse(strings.NewReader("name: test\nunknown: value\nsteps: []\n"))
	if err == nil {
		t.Fatal("expected unknown field error")
	}
}
