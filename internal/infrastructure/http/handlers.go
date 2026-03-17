package httpapi

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	appInvoice "github.com/rcarvalho-pb/payment_project-go/internal/application/invoice"
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/invoice"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/observability"
)

type InvoiceHandler struct {
	service   *appInvoice.Service
	templates *template.Template
}

type CreateInvoice struct {
	ID     string `json:"id"`
	Amount int64  `json:"amount"`
}

func NewInvoiceHandler(service *appInvoice.Service) *InvoiceHandler {
	funcMap := template.FuncMap{
		"lower": strings.ToLower,
	}
	tmpl := template.Must(
		template.New("").Funcs(funcMap).
			ParseGlob("internal/infrastructure/http/views/*.html"),
	)
	return &InvoiceHandler{
		service:   service,
		templates: tmpl,
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
	invoices, err := h.service.ListInvoices()
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

func (h *InvoiceHandler) Index(w http.ResponseWriter, r *http.Request) {
	err := h.templates.ExecuteTemplate(w, "layout.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *InvoiceHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	invoices, err := h.service.ListInvoices()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.templates.ExecuteTemplate(w, "invoices_table.html", invoices)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *InvoiceHandler) CreateInvoiceWeb(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	amountStr := r.FormValue("amount")
	amount, _ := strconv.Atoi(amountStr)
	if amount == 0 {
		amount = 1000 // valor default para demo
	}

	_, err := h.service.CreateInvoice(id, int64(amount))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.templates.ExecuteTemplate(w, "alert.html", err.Error())
		return
	}

	h.ListInvoices(w, r)
}

func (h *InvoiceHandler) PayInvoice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	cid := uuid.NewString()
	ctx := observability.WithCorrelationID(r.Context(), cid)

	err := h.service.RequestPayment(ctx, id)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		h.templates.ExecuteTemplate(w, "alert.html", err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *InvoiceHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	details, err := h.service.Repo.FindByID(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.templates.ExecuteTemplate(w, "invoice_details.html", details)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *InvoiceHandler) InvoiceRow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	invoice, err := h.service.Repo.FindByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.templates.ExecuteTemplate(w, "invoice_row.html", invoice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
