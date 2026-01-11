package invoice

import (
	"errors"
	"time"

	"github.com/rcarvalho-pb/payment_project-go/internal/application/worker"
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/invoice"
)

var (
	ErrInvoiceNotFound     = errors.New("invoice not found")
	ErrInvalidInvoiceState = errors.New("invalid invoice state")
)

type Service struct {
	Repo     invoice.Repository
	Recorder worker.Recorder
}

type EventPublisher interface {
	Publish(*event.Event) error
}

func (s *Service) CreateInvoice(id string, amount int64) (*invoice.Invoice, error) {
	inv := invoice.NewInvoice(id, amount)
	if err := s.Repo.Save(inv); err != nil {
		return nil, ErrInvoiceNotFound
	}

	return inv, nil
}

func (s *Service) RequestPayment(invoiceID string) error {
	inv, err := s.Repo.FindByID(invoiceID)
	if err != nil {
		return err
	}
	if inv.Status != invoice.StatusPending {
		return ErrInvalidInvoiceState
	}

	if err := s.Repo.UpdateStatus(inv.ID, invoice.StatusProcessing); err != nil {
		return err
	}

	evt := &event.Event{
		Type: event.PaymentRequested,
		Payload: &event.PaymentRequestPayload{
			InvoiceID: invoiceID,
			Amount:    inv.Amount,
			Attempt:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	return s.Recorder.Record(evt)
}
