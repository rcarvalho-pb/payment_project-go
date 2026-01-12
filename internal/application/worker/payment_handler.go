package worker

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
	domainPayment "github.com/rcarvalho-pb/payment_project-go/internal/domain/payment"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/logging"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/metrics"
)

var ErrInvalidPayload = errors.New("invalid payload for payment request")

type PaymentProcessor struct {
	Repo            domainPayment.Repository
	Recorder        Recorder
	Retry           *RetryScheduler
	PaymentExecutor PaymentExecutor
	Logger          logging.Logger
	Metrics         metrics.Counters
}

func (p *PaymentProcessor) Handle(evt *event.Event) error {
	if evt.Type != event.PaymentRequested {
		return nil
	}

	payload, ok := evt.Payload.(*event.PaymentRequestPayload)
	if !ok {
		return ErrInvalidPayload
	}

	p.Logger.Info("processing payment", map[string]any{
		"invoice-id": payload.InvoiceID,
		"attempt":    payload.Attempt,
	})

	idempotencyKey := uuid.NewString()

	_, err := p.Repo.FindByIdempotencyKey(idempotencyKey)
	if err == nil {
		return nil
	}

	paymentID := uuid.NewString()

	paymnt := domainPayment.NewPayment(paymentID, payload.InvoiceID, idempotencyKey)

	saved, err := p.Repo.SaveIfNotExist(paymnt)
	if err != nil && !saved {
		return err
	}

	paymentSucceeded := p.PaymentExecutor.Execute()

	p.Metrics.IncProcessed()

	if paymentSucceeded {
		p.Metrics.IncSucceeded()

		p.Logger.Info("payment succeeded", map[string]any{
			"payment-id": paymentID,
			"invoice-id": payload.InvoiceID,
			"attempt":    payload.Attempt,
		})

		if err := p.Repo.UpdateStatus(paymentID, domainPayment.StatusSuccess); err != nil {
			return err
		}

		return p.Recorder.Record(&event.Event{
			Type: event.PaymentSucceeded,
			Payload: &event.PaymentSucceededPayload{
				InvoiceID:  payload.InvoiceID,
				PaymentID:  paymentID,
				FinishedAt: time.Now(),
			},
		})
	}

	p.Logger.Info("payment failed", map[string]any{
		"payment-id": paymentID,
		"invoice-id": payload.InvoiceID,
		"attempt":    payload.Attempt,
	})

	p.Repo.UpdateStatus(paymentID, domainPayment.StatusFailed)

	p.Recorder.Record(&event.Event{
		Type: event.PaymentFailed,
		Payload: &event.PaymentFailedPayload{
			InvoiceID:  payload.InvoiceID,
			PaymentID:  paymentID,
			Retryable:  true,
			Reason:     "temporary failure",
			FinishedAt: time.Now(),
		},
	})

	p.Retry.ScheduleRetry(payload)

	return nil
}
