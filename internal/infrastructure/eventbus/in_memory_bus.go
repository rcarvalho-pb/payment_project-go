package eventbus

import (
	"sync"

	"github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
)

type HandlerFunc func(*event.Event) error

type InMemoryBus struct {
	mu       sync.RWMutex
	handlers map[event.Type][]HandlerFunc
}

func NewInMemoryBus() *InMemoryBus {
	return &InMemoryBus{
		mu:       sync.RWMutex{},
		handlers: make(map[event.Type][]HandlerFunc),
	}
}

func (b *InMemoryBus) Subscribe(eventType event.Type, handler HandlerFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *InMemoryBus) Publish(evt *event.Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, handler := range b.handlers[evt.Type] {
		err := handler(evt)
		if err != nil {
			return err
		}
	}
	return nil
}
