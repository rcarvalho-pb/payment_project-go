package web_handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/rcarvalho-pb/payment_project-go/internal/application/contracts"
	"github.com/rcarvalho-pb/payment_project-go/internal/application/invoice"
	domain "github.com/rcarvalho-pb/payment_project-go/internal/domain/invoice"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/observability"
	"github.com/rcarvalho-pb/payment_project-go/internal/views"
	"github.com/rcarvalho-pb/payment_project-go/internal/views/components"
)

type WebHandler struct {
	service *invoice.Service
	metrics contracts.PaymentMetrics
}

func NewWebHandler(service *invoice.Service, metrics contracts.PaymentMetrics) *WebHandler {
	return &WebHandler{
		service: service,
		metrics: metrics,
	}
}

func (h *WebHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	invoices, err := h.service.ListInvoices()
	if err != nil {
		log.Println("error getting invoices")
	}
	initialMetrics := components.ChartData{
		Labels: []string{"pending", "processed", "succeeded", "failed"},
		Values: []uint64{
			// Exemplo de como calcular ou buscar os valores atuais
			h.metrics.Pending(),
			h.metrics.Processed(),
			h.metrics.Succeeded(),
			h.metrics.Failed(),
		},
	}
	views.Layout(invoices, initialMetrics).Render(r.Context(), w)
}

func (h *WebHandler) HandleNewInvoice(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	amount, err := strconv.ParseInt(r.FormValue("amount"), 10, 10)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(id, amount)

	inv, err := h.service.CreateInvoice(id, int64(amount))
	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("HX-Trigger", "update-charts")

	components.InvoiceRow(inv).Render(r.Context(), w)
}

func (h *WebHandler) HandlePayment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	inv, err := h.service.Repo.FindByID(id)
	if err != nil {
		http.Error(w, "invalid id", http.StatusNotFound)
		return
	}

	cid := uuid.NewString()
	ctx := observability.WithCorrelationID(r.Context(), cid)

	if err := h.service.RequestPayment(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	status, err := h.service.Repo.GetStatus(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	inv.Status = domain.Status(status)

	w.Header().Set("HX-Trigger", "update-charts")

	components.InvoiceRow(inv).Render(r.Context(), w)
}

func (h *WebHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	inv, err := h.service.Repo.FindByID(id)
	if err != nil {
		log.Println(err)
	}

	components.InvoiceRow(inv).Render(r.Context(), w)
}

func (h *WebHandler) HandleLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, _ := w.(http.Flusher)

	// Exemplo de stream: em um projeto real, você usaria um channel
	for i := 0; i < 100; i++ {
		fmt.Fprintf(w, "data: <div class='log-entry'>[%s] Evento de log %d</div>\n\n",
			time.Now().Format("15:04:05"), i)
		flusher.Flush()
		time.Sleep(2 * time.Second)
	}
}

func (h *WebHandler) HandleDetails(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	inv, err := h.service.Repo.FindByID(id)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusNotFound)
		return
	}

	components.ItemDetailsModal(inv).Render(r.Context(), w)
}

func (h *WebHandler) HandleMetricsUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data := components.ChartData{
		Labels: []string{"pending", "processed", "succeeded", "failed"},
		Values: []uint64{
			// Exemplo de como calcular ou buscar os valores atuais
			h.metrics.Pending(),
			h.metrics.Processed(),
			h.metrics.Succeeded(),
			h.metrics.Failed(),
		},
	}
	json.NewEncoder(w).Encode(data)
}
