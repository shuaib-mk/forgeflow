package plugins

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sync"
)

var versionPattern = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

type Manifest struct { Name, Version, Description string }
type StepInput struct { WorkingDirectory string; Args []string }
type StepOutput struct { Logs string; Metadata map[string]string }
type StepHandler func(context.Context, StepInput) (StepOutput, error)
type NotificationHandler func(context.Context, string, map[string]any) error
type AnalyticsProcessor func(context.Context, map[string]float64) (map[string]float64, error)
type RepositoryIntegration interface { Sync(context.Context, string) error }

type Plugin interface {
	Manifest() Manifest
	Register(*Registrar) error
	Close(context.Context) error
}

type Registrar struct {
	steps map[string]StepHandler
	notifications map[string]NotificationHandler
	analytics map[string]AnalyticsProcessor
	repositories map[string]RepositoryIntegration
}

func (r *Registrar) WorkflowStep(name string, handler StepHandler) error {
	if name == "" || handler == nil { return errors.New("plugin step requires a name and handler") }
	if _, exists := r.steps[name]; exists { return fmt.Errorf("plugin step %q already registered", name) }
	r.steps[name] = handler
	return nil
}

func (r *Registrar) Notification(name string, handler NotificationHandler) error {
	if name == "" || handler == nil { return errors.New("notification handler requires a name and implementation") }
	if _, exists := r.notifications[name]; exists { return fmt.Errorf("notification handler %q already registered", name) }
	r.notifications[name] = handler
	return nil
}

func (r *Registrar) Analytics(name string, processor AnalyticsProcessor) error {
	if name == "" || processor == nil { return errors.New("analytics processor requires a name and implementation") }
	if _, exists := r.analytics[name]; exists { return fmt.Errorf("analytics processor %q already registered", name) }
	r.analytics[name] = processor
	return nil
}

func (r *Registrar) Repository(name string, integration RepositoryIntegration) error {
	if name == "" || integration == nil { return errors.New("repository integration requires a name and implementation") }
	if _, exists := r.repositories[name]; exists { return fmt.Errorf("repository integration %q already registered", name) }
	r.repositories[name] = integration
	return nil
}

type Registry struct {
	mu sync.RWMutex
	plugins map[string]Plugin
	steps map[string]StepHandler
}

func NewRegistry() *Registry { return &Registry{plugins: map[string]Plugin{}, steps: map[string]StepHandler{}} }

func (r *Registry) Register(plugin Plugin) error {
	if plugin == nil { return errors.New("plugin is nil") }
	manifest := plugin.Manifest()
	if manifest.Name == "" || !versionPattern.MatchString(manifest.Version) {
		return errors.New("plugin manifest requires a name and semantic version prefixed with v")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.plugins[manifest.Name]; exists { return fmt.Errorf("plugin %q already registered", manifest.Name) }
	registrar := &Registrar{steps: map[string]StepHandler{}, notifications: map[string]NotificationHandler{}, analytics: map[string]AnalyticsProcessor{}, repositories: map[string]RepositoryIntegration{}}
	if err := plugin.Register(registrar); err != nil { return fmt.Errorf("register plugin %q: %w", manifest.Name, err) }
	for name := range registrar.steps {
		if _, exists := r.steps[name]; exists { return fmt.Errorf("workflow step %q conflicts with another plugin", name) }
	}
	r.plugins[manifest.Name] = plugin
	for name, handler := range registrar.steps { r.steps[name] = handler }
	return nil
}

func (r *Registry) Step(name string) (StepHandler, bool) {
	r.mu.RLock(); defer r.mu.RUnlock()
	handler, ok := r.steps[name]
	return handler, ok
}

func (r *Registry) Close(ctx context.Context) error {
	r.mu.RLock(); defer r.mu.RUnlock()
	var errs []error
	for name, plugin := range r.plugins {
		if err := plugin.Close(ctx); err != nil { errs = append(errs, fmt.Errorf("close %s: %w", name, err)) }
	}
	return errors.Join(errs...)
}
