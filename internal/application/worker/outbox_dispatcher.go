package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/logging"
	"github.com/rcarvalho-pb/payment_project-go/internal/infrastructure/outbox"
)

type OutboxDispatcher struct {
	Repo         outbox.OutboxRepository
	EventBus     EventPublisher
	Logger       logging.Logger
	PollInterval time.Duration
	BatchSize    int
}

func (d *OutboxDispatcher) Run(ctx context.Context) {
	ticker := time.NewTicker(d.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.dispatchOnce()
		}
	}
}

func (d *OutboxDispatcher) dispatchOnce() {
	events, err := d.Repo.FindUnpublished(d.BatchSize)
	if err != nil {
		d.Logger.Error("error finding unpublished outbox events", nil)
		return
	}
	for _, evt := range events {
		var payload any
		if err := json.Unmarshal(evt.Payload, &payload); err != nil {
			continue
		}

		domainEvent := &event.Event{
			Type:    evt.Type,
			Payload: payload,
		}

		if err := d.EventBus.Publish(domainEvent); err != nil {
			d.Logger.Error("error publishing event in eventbus", nil)
			continue
		}

		if err := d.Repo.MarkPublished(evt.ID); err != nil {
			d.Logger.Error("error marking outbox event as published", nil)
			continue
		}
	}
}
