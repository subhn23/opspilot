package events

import "sync"

type EventBus struct {
	mu        sync.RWMutex
	listeners []chan bool
}

var GlobalBus = &EventBus{}

func (b *EventBus) Subscribe() chan bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan bool, 1)
	b.listeners = append(b.listeners, ch)
	return ch
}

func (b *EventBus) Unsubscribe(ch chan bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, listener := range b.listeners {
		if listener == ch {
			b.listeners = append(b.listeners[:i], b.listeners[i+1:]...)
			break
		}
	}
}

func (b *EventBus) Publish() {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.listeners {
		select {
		case ch <- true:
		default:
		}
	}
}

func Notify() {
	GlobalBus.Publish()
}
