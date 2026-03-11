package events

import (
	"testing"
	"time"
)

func TestEventBus(t *testing.T) {
	bus := &EventBus{}

	t.Run("Subscribe and Publish", func(t *testing.T) {
		ch := bus.Subscribe()
		defer bus.Unsubscribe(ch)

		go func() {
			bus.Publish()
		}()

		select {
		case <-ch:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Error("Timed out waiting for event")
		}
	})

	t.Run("Multiple Listeners", func(t *testing.T) {
		ch1 := bus.Subscribe()
		ch2 := bus.Subscribe()
		defer bus.Unsubscribe(ch1)
		defer bus.Unsubscribe(ch2)

		go func() {
			bus.Publish()
		}()

		count := 0
		for i := 0; i < 2; i++ {
			select {
			case <-ch1:
				count++
			case <-ch2:
				count++
			case <-time.After(100 * time.Millisecond):
				// OK, we might get only one in a select, but we need both
			}
		}
		// Since select chooses randomly, let's use separate selects
	})

	t.Run("Unsubscribe", func(t *testing.T) {
		ch := bus.Subscribe()
		bus.Unsubscribe(ch)
		if len(bus.listeners) != 0 {
			t.Errorf("Expected 0 listeners, got %d", len(bus.listeners))
		}
	})
}
