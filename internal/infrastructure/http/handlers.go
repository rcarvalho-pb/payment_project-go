package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	appInvoice "github.com/rcarvalho-pb/payment_project-go/internal/application/invoice"
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/invoice"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/observability"
)

type InvoiceHandler struct {
	service *appInvoice.Service
}

type CreateInvoice struct {
	ID     string `json:"id"`
	Amount int64  `json:"amount"`
}

func NewInvoiceHandler(service *appInvoice.Service) *InvoiceHandler {
	return &InvoiceHandler{
		service: service,
	}
}

func (h *InvoiceHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	var req CreateInvoice

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	inv, err := h.service.CreateInvoice(req.ID, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(inv)
}

func (h *InvoiceHandler) RequestPayment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	cid := uuid.NewString()
	ctx := observability.WithCorrelationID(r.Context(), cid)

	if err := h.service.RequestPayment(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *InvoiceHandler) GetInvoices(w http.ResponseWriter, r *http.Request) {
	invoices, err := h.service.Repo.FindAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	invoicesDTO := make([]invoice.InvoiceDTO, 0)

	for i := range invoices {
		invoicesDTO = append(invoicesDTO, invoice.InvoiceDTO{
			ID:        invoices[i].ID,
			Amount:    invoices[i].Amount,
			Status:    invoices[i].Status.String(),
			CreatedAt: invoices[i].CreatedAt,
			UpdatedAt: invoices[i].UpdatedAt,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&invoicesDTO)
}
