package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/forgeflow/forgeflow/pkg/models"
)

type Event interface {
	Name() string
	OccurredAt() time.Time
}

type ProjectCreated struct {
	ProjectID models.ID
	ActorID   models.ID
	At        time.Time
}

func (e ProjectCreated) Name() string          { return "project.created" }
func (e ProjectCreated) OccurredAt() time.Time { return e.At }

type WorkflowCompleted struct {
	RunID     models.ID
	ProjectID models.ID
	Status    models.RunStatus
	At        time.Time
}

func (e WorkflowCompleted) Name() string          { return "workflow.completed" }
func (e WorkflowCompleted) OccurredAt() time.Time { return e.At }

type Handler func(context.Context, Event) error

type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

func NewBus() *Bus { return &Bus{handlers: make(map[string][]Handler)} }

func (b *Bus) Subscribe(name string, handler Handler) func() {
	b.mu.Lock()
	b.handlers[name] = append(b.handlers[name], handler)
	index := len(b.handlers[name]) - 1
	b.mu.Unlock()
	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if current := b.handlers[name]; index < len(current) && current[index] != nil {
			current[index] = nil
		}
	}
}

func (b *Bus) Publish(ctx context.Context, event Event) error {
	b.mu.RLock()
	handlers := append([]Handler(nil), b.handlers[event.Name()]...)
	b.mu.RUnlock()
	for _, handler := range handlers {
		if handler == nil {
			continue
		}
		if err := handler(ctx, event); err != nil {
			return fmt.Errorf("handle %s: %w", event.Name(), err)
		}
	}
	return nil
}
