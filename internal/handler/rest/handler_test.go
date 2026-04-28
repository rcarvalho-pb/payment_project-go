package rest_handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	appInvoice "github.com/rcarvalho-pb/payment_project-go/internal/application/invoice"
	"github.com/rcarvalho-pb/payment_project-go/internal/application/worker"
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
	domainInvoice "github.com/rcarvalho-pb/payment_project-go/internal/domain/invoice"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/metrics"
)

type restRepoStub struct {
	saveErr error
	findErr error
	invoice *domainInvoice.Invoice
}

func (r *restRepoStub) Save(*domainInvoice.Invoice) error { return r.saveErr }
func (r *restRepoStub) FindByID(string) (*domainInvoice.Invoice, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	return r.invoice, nil
}
func (r *restRepoStub) FindAll() ([]*domainInvoice.Invoice, error) { return nil, nil }
func (r *restRepoStub) GetStatus(string) (uint8, error) {
	return uint8(domainInvoice.StatusPending), nil
}
func (r *restRepoStub) UpdateStatus(string, domainInvoice.Status) error {
	return nil
}
func (r *restRepoStub) UpdateStatusTx(*sql.Tx, string, domainInvoice.Status) error {
	return nil
}

type restRecorderStub struct{}

func (r *restRecorderStub) Record(context.Context, *event.Event) error { return nil }
func (r *restRecorderStub) RecordTx(context.Context, *sql.Tx, *event.Event) error {
	return nil
}

func TestCreateInvoiceReturnsJSONValidationError(t *testing.T) {
	service := &appInvoice.Service{
		Repo:     &restRepoStub{},
		Recorder: worker.Recorder(&restRecorderStub{}),
		Metrics:  &metrics.Counters{},
	}
	handler := NewRestHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/api/invoices", bytes.NewBufferString(`{"id":"","amount":0}`))
	res := httptest.NewRecorder()

	handler.CreateInvoice(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}

	var payload errorResponse
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload.Error.Code != "invalid_invoice_id" {
		t.Fatalf("expected invalid_invoice_id, got %s", payload.Error.Code)
	}
}

func TestRequestPaymentReturnsConflictJSON(t *testing.T) {
	service := &appInvoice.Service{
		Repo: &restRepoStub{
			invoice: &domainInvoice.Invoice{
				ID:     "inv-1",
				Amount: 100,
				Status: domainInvoice.StatusPaid,
			},
		},
		Recorder: worker.Recorder(&restRecorderStub{}),
		Metrics:  &metrics.Counters{},
	}
	handler := NewRestHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/api/invoices/inv-1/pay", nil)
	req.SetPathValue("id", "inv-1")
	res := httptest.NewRecorder()

	handler.RequestPayment(res, req)

	if res.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", res.Code)
	}

	var payload errorResponse
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload.Error.Code != "invalid_invoice_state" {
		t.Fatalf("expected invalid_invoice_state, got %s", payload.Error.Code)
	}
}
