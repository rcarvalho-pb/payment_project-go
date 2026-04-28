package invoice

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/rcarvalho-pb/payment_project-go/internal/application/worker"
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
	domainInvoice "github.com/rcarvalho-pb/payment_project-go/internal/domain/invoice"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/metrics"
)

type fakeInvoiceRepo struct {
	invoice              *domainInvoice.Invoice
	updateStatusCalled   bool
	updateStatusTxCalled bool
	saveErr              error
}

func (f *fakeInvoiceRepo) Save(*domainInvoice.Invoice) error { return f.saveErr }

func (f *fakeInvoiceRepo) FindByID(string) (*domainInvoice.Invoice, error) {
	return f.invoice, nil
}

func (f *fakeInvoiceRepo) FindAll() ([]*domainInvoice.Invoice, error) { return nil, nil }

func (f *fakeInvoiceRepo) GetStatus(string) (uint8, error) { return uint8(f.invoice.Status), nil }

func (f *fakeInvoiceRepo) UpdateStatus(string, domainInvoice.Status) error {
	f.updateStatusCalled = true
	return nil
}

func (f *fakeInvoiceRepo) UpdateStatusTx(*sql.Tx, string, domainInvoice.Status) error {
	f.updateStatusTxCalled = true
	return nil
}

type failingRecorder struct{}

func (f *failingRecorder) Record(context.Context, *event.Event) error {
	return errors.New("record failed")
}

func (f *failingRecorder) RecordTx(context.Context, *sql.Tx, *event.Event) error {
	return errors.New("record failed")
}

type testUOW struct {
	db *sql.DB
}

func (u *testUOW) Begin() (*sql.Tx, error) {
	return u.db.Begin()
}

func TestRequestPaymentUsesTransactionAndRollsBackWhenRecordingFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mock.ExpectBegin()
	mock.ExpectRollback()

	repo := &fakeInvoiceRepo{
		invoice: &domainInvoice.Invoice{
			ID:        "inv-1",
			Amount:    1500,
			Status:    domainInvoice.StatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			DueDate:   time.Now().Add(time.Hour),
		},
	}
	var recorder worker.Recorder = &failingRecorder{}
	service := &Service{
		Repo:     repo,
		Recorder: recorder,
		Metrics:  &metrics.Counters{},
		UOW:      &testUOW{db: db},
	}

	err = service.RequestPayment(context.Background(), repo.invoice.ID)
	if err == nil {
		t.Fatal("expected request payment to fail")
	}
	if !repo.updateStatusTxCalled {
		t.Fatal("expected transactional status update to be used")
	}
	if repo.updateStatusCalled {
		t.Fatal("did not expect non-transactional status update")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestCreateInvoiceValidatesInput(t *testing.T) {
	service := &Service{
		Repo:    &fakeInvoiceRepo{},
		Metrics: &metrics.Counters{},
	}

	if _, err := service.CreateInvoice("", 100); !errors.Is(err, ErrInvalidInvoiceID) {
		t.Fatalf("expected invalid invoice id error, got %v", err)
	}

	if _, err := service.CreateInvoice("inv-1", 0); !errors.Is(err, ErrInvalidInvoiceAmount) {
		t.Fatalf("expected invalid invoice amount error, got %v", err)
	}
}

func TestCreateInvoiceMapsDuplicateKeyToDomainError(t *testing.T) {
	service := &Service{
		Repo: &fakeInvoiceRepo{
			saveErr: errors.New("UNIQUE constraint failed: invoices.id"),
		},
		Metrics: &metrics.Counters{},
	}

	_, err := service.CreateInvoice("inv-1", 100)
	if !errors.Is(err, ErrInvoiceAlreadyExists) {
		t.Fatalf("expected duplicate key to map to ErrInvoiceAlreadyExists, got %v", err)
	}
}
