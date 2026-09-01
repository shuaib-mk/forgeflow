package plugins

import (
	"context"
	"testing"
)

type testPlugin struct{ name string }

func (p testPlugin) Manifest() Manifest { return Manifest{Name: p.name, Version: "v1.0.0"} }
func (p testPlugin) Register(r *Registrar) error {
	return r.WorkflowStep("echo", func(context.Context, StepInput) (StepOutput, error) { return StepOutput{Logs: "ok"}, nil })
}
func (p testPlugin) Close(context.Context) error { return nil }

func TestRegistryRejectsDuplicatePlugin(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	if err := r.Register(testPlugin{name: "example"}); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(testPlugin{name: "example"}); err == nil {
		t.Fatal("expected duplicate error")
	}
}
