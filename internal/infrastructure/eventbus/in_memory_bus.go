package eventbus

import (
	"context"
	"sync"

	"github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/invoice"
)

type HandlerFunc func(context.Context, *event.Event) error

type InMemoryBus struct {
	invoiceRepo invoice.Repository
	mu          sync.RWMutex
	handlers    map[event.Type][]HandlerFunc
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

func (b *InMemoryBus) Publish(ctx context.Context, evt *event.Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, handler := range b.handlers[evt.Type] {
		err := handler(ctx, evt)
		if err != nil {
			return err
		}
	}
	return nil
}
