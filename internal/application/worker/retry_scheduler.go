package worker

import (
	"time"

	"github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_project-go/internal/infrastructure/outbox"
)

type RetryScheduler struct {
	Recorder  outbox.Recorder
	MaxRetry  int
	BaseDelay time.Duration
	MaxDelay  time.Duration
}

func (r *RetryScheduler) ScheduleRetry(payload *event.PaymentRequestPayload) {
	if payload.Attempt >= r.MaxRetry {
		return
	}

	delay := r.BaseDelay * time.Duration(1<<payload.Attempt-1)
	if delay > r.MaxDelay {
		delay = r.MaxDelay
	}

	nextPayload := &event.PaymentRequestPayload{
		InvoiceID: payload.InvoiceID,
		Amount:    payload.Amount,
		Attempt:   payload.Attempt + 1,
		CreatedAt: payload.CreatedAt,
		UpdatedAt: time.Now(),
	}

	go func() {
		time.Sleep(delay)
		r.Recorder.Record(&event.Event{
			Type:    event.PaymentRequested,
			Payload: nextPayload,
		})
	}()
}
