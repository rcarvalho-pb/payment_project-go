package invoice

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/rcarvalho-pb/payment_project-go/internal/application/contracts"
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
	Metrics  contracts.PaymentMetrics
	UOW      interface {
		Begin() (*sql.Tx, error)
	}
}

func (s *Service) CreateInvoice(id string, amount int64) (*invoice.Invoice, error) {
	inv := invoice.NewInvoice(id, amount)
	if err := s.Repo.Save(inv); err != nil {
		return nil, ErrInvoiceNotFound
	}

	s.Metrics.IncPending()

	return inv, nil
}

func (s *Service) RequestPayment(ctx context.Context, invoiceID string) error {
	inv, err := s.Repo.FindByID(invoiceID)
	if err != nil {
		return err
	}
	if inv.Status != invoice.StatusPending {
		return ErrInvalidInvoiceState
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

	if s.UOW != nil {
		repo, ok := s.Repo.(interface {
			UpdateStatusTx(*sql.Tx, string, invoice.Status) error
		})
		recorder, recorderOK := s.Recorder.(interface {
			RecordTx(context.Context, *sql.Tx, *event.Event) error
		})
		if ok && recorderOK {
			tx, err := s.UOW.Begin()
			if err != nil {
				return err
			}
			defer tx.Rollback()

			if err := repo.UpdateStatusTx(tx, inv.ID, invoice.StatusProcessing); err != nil {
				return err
			}

			if err := recorder.RecordTx(ctx, tx, evt); err != nil {
				return err
			}

			return tx.Commit()
		}
	}

	if err := s.Repo.UpdateStatus(inv.ID, invoice.StatusProcessing); err != nil {
		return err
	}

	return s.Recorder.Record(ctx, evt)
}

func (s *Service) ListInvoices() ([]*invoice.Invoice, error) {
	return s.Repo.FindAll()
}
