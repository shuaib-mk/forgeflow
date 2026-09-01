package events

import (
	"context"
	"testing"
	"time"

	"github.com/forgeflow/forgeflow/pkg/models"
)

func TestBusSubscribePublishAndUnsubscribe(t *testing.T) {
	t.Parallel()
	bus := NewBus()
	calls := 0
	unsubscribe := bus.Subscribe("project.created", func(_ context.Context, event Event) error {
		calls++
		if event.(ProjectCreated).ProjectID != "p1" {
			t.Fatal("unexpected project")
		}
		return nil
	})
	event := ProjectCreated{ProjectID: models.ID("p1"), At: time.Now()}
	if err := bus.Publish(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	unsubscribe()
	if err := bus.Publish(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}
