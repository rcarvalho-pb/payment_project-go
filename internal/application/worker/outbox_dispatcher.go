package worker

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/rcarvalho-pb/payment_project-go/internal/application/contracts"
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/logging"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/observability"
	"github.com/rcarvalho-pb/payment_project-go/internal/infrastructure/outbox"
)

type OutboxDispatcher struct {
	Repo         outbox.OutboxRepository
	EventBus     event.EventPublisher
	Metrics      contracts.OutboxMetrics
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
			d.dispatchOnce(ctx)
		}
	}
}

func (d *OutboxDispatcher) dispatchOnce(ctx context.Context) {
	events, ids, err := d.Repo.FindUnpublished(d.BatchSize)
	if err != nil {
		d.Logger.Error("error finding unpublished outbox events: "+err.Error(), nil)
		return
	}
	for _, evt := range events {
		payload, err := unmarshallPayload(evt)
		if err != nil || payload == nil {
			d.Logger.Error("error unmarshiling payload: "+err.Error(), nil)
			continue
		}

		domainEvent := &event.Event{
			Type:    evt.Type,
			Payload: payload,
		}

		if evt.CorrelationID != "" {
			ctx = observability.WithCorrelationID(ctx, evt.CorrelationID)
		}

		if err := d.EventBus.Publish(ctx, domainEvent); err != nil {
			d.Metrics.IncPublishFailed()
			d.Logger.Error("error publishing event in eventbus: "+err.Error(), nil)
			continue
		}
		d.Metrics.IncPublished()
	}
	if err := d.Repo.MarkPublished(ids); err != nil {
		d.Logger.Error("error marking outbox event as published", nil)
		return
	}
}

func unmarshallPayload(evt *outbox.OutboxEvent) (any, error) {
	switch evt.Type {
	case event.PaymentRequested:
		var payload event.PaymentRequestPayload
		if err := json.Unmarshal(evt.Payload, &payload); err != nil {
			return nil, err
		}
		return &payload, nil

	case event.PaymentSucceeded:
		var payload event.PaymentSucceededPayload
		if err := json.Unmarshal(evt.Payload, &payload); err != nil {
			return nil, err
		}
		return &payload, nil

	case event.PaymentFailed:
		var payload event.PaymentFailedPayload
		if err := json.Unmarshal(evt.Payload, &payload); err != nil {
			return nil, err
		}
		return &payload, nil
	}
	return nil, errors.New("invalid event type")
}
