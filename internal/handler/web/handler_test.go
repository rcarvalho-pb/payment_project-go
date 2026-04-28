package web_handler

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	appInvoice "github.com/rcarvalho-pb/payment_project-go/internal/application/invoice"
	"github.com/rcarvalho-pb/payment_project-go/internal/application/worker"
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
	domainInvoice "github.com/rcarvalho-pb/payment_project-go/internal/domain/invoice"
	domainPayment "github.com/rcarvalho-pb/payment_project-go/internal/domain/payment"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/logging"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/metrics"
)

type webInvoiceRepoStub struct {
	saveErr error
	findErr error
	invoice *domainInvoice.Invoice
}

func (r *webInvoiceRepoStub) Save(*domainInvoice.Invoice) error { return r.saveErr }
func (r *webInvoiceRepoStub) FindByID(string) (*domainInvoice.Invoice, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	return r.invoice, nil
}
func (r *webInvoiceRepoStub) FindAll() ([]*domainInvoice.Invoice, error) { return nil, nil }
func (r *webInvoiceRepoStub) GetStatus(string) (uint8, error)            { return uint8(r.invoice.Status), nil }
func (r *webInvoiceRepoStub) UpdateStatus(string, domainInvoice.Status) error {
	return nil
}
func (r *webInvoiceRepoStub) UpdateStatusTx(*sql.Tx, string, domainInvoice.Status) error {
	return nil
}

type webPaymentRepoStub struct{}

func (r *webPaymentRepoStub) SaveIfNotExist(*domainPayment.Payment) (bool, error) { return true, nil }
func (r *webPaymentRepoStub) FindByIdempotencyKey(string) (*domainPayment.Payment, error) {
	return nil, nil
}
func (r *webPaymentRepoStub) FindByInvoiceID(string) ([]*domainPayment.Payment, error) {
	return nil, nil
}
func (r *webPaymentRepoStub) FindAll() ([]*domainPayment.Payment, error)      { return nil, nil }
func (r *webPaymentRepoStub) UpdateStatus(string, domainPayment.Status) error { return nil }

type webRecorderStub struct{}

func (r *webRecorderStub) Record(context.Context, *event.Event) error { return nil }
func (r *webRecorderStub) RecordTx(context.Context, *sql.Tx, *event.Event) error {
	return nil
}

func TestHandleNewInvoiceReturnsFeedbackHeaderOnValidationError(t *testing.T) {
	service := &appInvoice.Service{
		Repo:     &webInvoiceRepoStub{},
		Recorder: worker.Recorder(&webRecorderStub{}),
		Metrics:  &metrics.Counters{},
	}
	logger := logging.NewStdoutLogger(nil)
	handler := NewWebHandler(service, &metrics.Counters{}, &webPaymentRepoStub{}, &logger, nil)

	req := httptest.NewRequest(http.MethodPost, "/invoices", nil)
	res := httptest.NewRecorder()

	handler.HandleNewInvoice(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}
	if res.Header().Get("HX-Trigger") == "" {
		t.Fatal("expected HX-Trigger feedback header")
	}
}

func TestHandlePaymentReturnsFeedbackHeaderOnInvalidState(t *testing.T) {
	service := &appInvoice.Service{
		Repo: &webInvoiceRepoStub{
			invoice: &domainInvoice.Invoice{
				ID:     "inv-1",
				Amount: 100,
				Status: domainInvoice.StatusPaid,
			},
		},
		Recorder: worker.Recorder(&webRecorderStub{}),
		Metrics:  &metrics.Counters{},
	}
	logger := logging.NewStdoutLogger(nil)
	handler := NewWebHandler(service, &metrics.Counters{}, &webPaymentRepoStub{}, &logger, nil)

	req := httptest.NewRequest(http.MethodPost, "/invoices/inv-1/pay", nil)
	req.SetPathValue("id", "inv-1")
	res := httptest.NewRecorder()

	handler.HandlePayment(res, req)

	if res.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", res.Code)
	}
	if res.Header().Get("HX-Trigger") == "" {
		t.Fatal("expected HX-Trigger feedback header")
	}
}

func TestHandleNewInvoiceAcceptsFourDigitAmount(t *testing.T) {
	service := &appInvoice.Service{
		Repo:     &webInvoiceRepoStub{},
		Recorder: worker.Recorder(&webRecorderStub{}),
		Metrics:  &metrics.Counters{},
	}
	logger := logging.NewStdoutLogger(nil)
	handler := NewWebHandler(service, &metrics.Counters{}, &webPaymentRepoStub{}, &logger, nil)

	form := url.Values{}
	form.Set("id", "inv-1000")
	form.Set("amount", "1500")

	req := httptest.NewRequest(http.MethodPost, "/invoices", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	handler.HandleNewInvoice(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	if !strings.Contains(res.Body.String(), "R$ 1500") {
		t.Fatalf("expected created row to contain four-digit amount, got %q", res.Body.String())
	}
}
