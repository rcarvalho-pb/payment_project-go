package rest_handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	appInvoice "github.com/rcarvalho-pb/payment_project-go/internal/application/invoice"
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/invoice"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/observability"
)

type RestHandler struct {
	service *appInvoice.Service
}

type CreateInvoice struct {
	ID     string `json:"id"`
	Amount int64  `json:"amount"`
}

func NewRestHandler(service *appInvoice.Service) *RestHandler {
	return &RestHandler{
		service: service,
	}
}

func (h *RestHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	var req CreateInvoice

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: apiError{
				Code:    "invalid_request_body",
				Message: "O corpo da requisição precisa conter um JSON válido.",
			},
		})
		return
	}

	inv, err := h.service.CreateInvoice(req.ID, req.Amount)
	if err != nil {
		writeAPIError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, inv)
}

func (h *RestHandler) RequestPayment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	cid := uuid.NewString()
	ctx := observability.WithCorrelationID(r.Context(), cid)

	if err := h.service.RequestPayment(ctx, id); err != nil {
		writeAPIError(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *RestHandler) GetInvoices(w http.ResponseWriter, r *http.Request) {
	invoices, err := h.service.Repo.FindAll()
	if err != nil {
		writeAPIError(w, err)
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

	writeJSON(w, http.StatusOK, &invoicesDTO)
}
