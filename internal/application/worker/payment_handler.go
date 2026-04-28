package worker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rcarvalho-pb/payment_project-go/internal/application/contracts"
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
	domainPayment "github.com/rcarvalho-pb/payment_project-go/internal/domain/payment"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/logging"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/observability"
)

var ErrInvalidPayload = errors.New("invalid payload for payment request")

type PaymentProcessor struct {
	Repo            domainPayment.Repository
	Recorder        Recorder
	Retry           *RetryScheduler
	PaymentExecutor PaymentExecutor
	Logger          logging.Logger
	Metrics         contracts.PaymentMetrics
	UOW             interface {
		Begin() (*sql.Tx, error)
	}
}

func (p *PaymentProcessor) Handle(ctx context.Context, evt *event.Event) error {
	if evt.Type != event.PaymentRequested {
		return nil
	}

	payload, ok := evt.Payload.(*event.PaymentRequestPayload)
	if !ok {
		return ErrInvalidPayload
	}

	fields := map[string]any{
		"invoice-id": payload.InvoiceID,
		"attempt":    payload.Attempt,
	}

	cid, ok := observability.CorrelationIDFromContext(ctx)
	if ok {
		fields["correlation-id"] = cid
	}

	p.Logger.Info("processing payment", fields)

	idempotencyKey := fmt.Sprintf("%s:%d", payload.InvoiceID, payload.Attempt)

	paymentID := uuid.NewString()

	fields["payment-id"] = paymentID

	paymnt := domainPayment.NewPayment(paymentID, payload.InvoiceID, idempotencyKey)
	paymnt.Attempt = payload.Attempt

	saved, err := p.Repo.SaveIfNotExist(paymnt)
	if err != nil {
		return err
	}
	if !saved {
		return nil
	}

	paymentSucceeded := p.PaymentExecutor.Execute()

	if payload.Attempt <= 1 {
		p.Metrics.DeincPending()
	}
	p.Metrics.IncProcessed()

	if paymentSucceeded {
		p.Metrics.IncSucceeded()

		p.Logger.Info("payment succeeded", fields)

		return p.persistResult(ctx, paymentID, domainPayment.StatusSuccess, &event.Event{
			Type: event.PaymentSucceeded,
			Payload: &event.PaymentSucceededPayload{
				InvoiceID:  payload.InvoiceID,
				PaymentID:  paymentID,
				FinishedAt: time.Now(),
			},
		})
	}

	p.Logger.Info("payment failed", fields)
	retryable := p.Retry != nil && p.Retry.CanRetry(payload.Attempt)
	failPayload := &event.PaymentFailedPayload{
		InvoiceID:  payload.InvoiceID,
		PaymentID:  paymentID,
		Retryable:  retryable,
		Reason:     "temporary failure",
		FinishedAt: time.Now(),
	}

	status := domainPayment.StatusTemporaryFailed
	if !retryable {
		status = domainPayment.StatusFailed
		failPayload.Reason = "max attempts reached"
		p.Metrics.IncFailed()
	}

	if err := p.persistResult(ctx, paymentID, status, &event.Event{
		Type:    event.PaymentFailed,
		Payload: failPayload,
	}); err != nil {
		return err
	}

	if !retryable {
		return nil
	}

	return p.Retry.ScheduleRetry(ctx, payload)
}

func (p *PaymentProcessor) persistResult(ctx context.Context, paymentID string, status domainPayment.Status, evt *event.Event) error {
	if p.UOW != nil {
		repo, ok := p.Repo.(interface {
			UpdateStatusTx(*sql.Tx, string, domainPayment.Status) error
		})
		recorder, recorderOK := p.Recorder.(interface {
			RecordTx(context.Context, *sql.Tx, *event.Event) error
		})
		if ok && recorderOK {
			tx, err := p.UOW.Begin()
			if err != nil {
				return err
			}
			defer tx.Rollback()

			if err := repo.UpdateStatusTx(tx, paymentID, status); err != nil {
				return err
			}

			if err := recorder.RecordTx(ctx, tx, evt); err != nil {
				return err
			}

			return tx.Commit()
		}
	}

	if err := p.Repo.UpdateStatus(paymentID, status); err != nil {
		return err
	}

	return p.Recorder.Record(ctx, evt)
}
