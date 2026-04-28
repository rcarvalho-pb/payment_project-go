package worker

import (
	"context"
	"errors"
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

func (r *RetryScheduler) CanRetry(attempt int) bool {
	return attempt < r.MaxRetry
}

func (r *RetryScheduler) ScheduleRetry(ctx context.Context, payload *event.PaymentRequestPayload) error {
	if !r.CanRetry(payload.Attempt) {
		return errors.New("max attempts reached")
	}

	delay := min(r.BaseDelay*time.Duration(1<<payload.Attempt-1), r.MaxDelay)

	nextPayload := event.PaymentRequestPayload{
		InvoiceID: payload.InvoiceID,
		Amount:    payload.Amount,
		Attempt:   payload.Attempt + 1,
		CreatedAt: payload.CreatedAt,
		UpdatedAt: time.Now(),
	}

	go func() {
		time.Sleep(delay)
		r.Recorder.Record(ctx, &event.Event{
			Type:    event.PaymentRequested,
			Payload: &nextPayload,
		})
	}()
	return nil
}
