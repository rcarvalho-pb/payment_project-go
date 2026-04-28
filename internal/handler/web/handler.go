package web_handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/rcarvalho-pb/payment_project-go/internal/application/contracts"
	"github.com/rcarvalho-pb/payment_project-go/internal/application/invoice"
	domain "github.com/rcarvalho-pb/payment_project-go/internal/domain/invoice"
	domainPayment "github.com/rcarvalho-pb/payment_project-go/internal/domain/payment"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/logging"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/observability"
	"github.com/rcarvalho-pb/payment_project-go/internal/views"
	"github.com/rcarvalho-pb/payment_project-go/internal/views/components"
)

type WebHandler struct {
	service   *invoice.Service
	metrics   contracts.PaymentMetrics
	payments  domainPayment.Repository
	logger    logging.Logger
	logStream *logging.Stream
}

func NewWebHandler(service *invoice.Service, metrics contracts.PaymentMetrics, payments domainPayment.Repository, logger logging.Logger, logStream *logging.Stream) *WebHandler {
	return &WebHandler{
		service:   service,
		metrics:   metrics,
		payments:  payments,
		logger:    logger,
		logStream: logStream,
	}
}

func (h *WebHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	invoices, err := h.service.ListInvoices()
	if err != nil {
		h.logger.Error("error getting invoices", map[string]any{"error": err.Error()})
	}
	initialMetrics := components.ChartData{
		Labels: []string{"pending", "succeeded", "failed"},
		Values: []uint64{
			// Exemplo de como calcular ou buscar os valores atuais
			h.metrics.Pending(),
			h.metrics.Succeeded(),
			h.metrics.Failed(),
		},
	}
	views.Layout(invoices, initialMetrics).Render(r.Context(), w)
}

func (h *WebHandler) HandleDashboardOverview(w http.ResponseWriter, r *http.Request) {
	invoices, err := h.service.ListInvoices()
	if err != nil {
		h.logger.Error("error getting invoices for dashboard overview", map[string]any{"error": err.Error()})
		http.Error(w, "Erro ao atualizar dashboard", http.StatusInternalServerError)
		return
	}

	views.DashboardOverview(views.BuildDashboardSummary(invoices)).Render(r.Context(), w)
}

func (h *WebHandler) HandleNewInvoice(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	amount, err := strconv.ParseInt(r.FormValue("amount"), 10, 64)
	if err != nil {
		h.logger.Error("invalid amount for invoice creation", map[string]any{"error": err.Error()})
		writeHTMXFeedback(w, http.StatusBadRequest, "danger", "Valor inválido", "Informe um valor numérico válido para a invoice.")
		return
	}

	h.logger.Info("creating invoice", map[string]any{"invoice-id": id, "amount": amount})

	inv, err := h.service.CreateInvoice(id, int64(amount))
	if err != nil {
		h.logger.Error("error creating invoice", map[string]any{"invoice-id": id, "error": err.Error()})
		status, title, message := mapInvoiceError(err)
		writeHTMXFeedback(w, status, "danger", title, message)
		return
	}

	writeSuccessHeaders(w, "Invoice criada", "A invoice foi criada e já está disponível para processamento.", true)

	components.InvoiceRow(inv).Render(r.Context(), w)
}

func (h *WebHandler) HandlePayment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	inv, err := h.service.Repo.FindByID(id)
	if err != nil {
		status, title, message := mapInvoiceError(err)
		writeHTMXFeedback(w, status, "danger", title, message)
		return
	}

	cid := uuid.NewString()
	ctx := observability.WithCorrelationID(r.Context(), cid)

	if err := h.service.RequestPayment(ctx, id); err != nil {
		status, title, message := mapInvoiceError(err)
		writeHTMXFeedback(w, status, "danger", title, message)
		return
	}

	status, err := h.service.Repo.GetStatus(id)
	if err != nil {
		writeHTMXFeedback(w, http.StatusInternalServerError, "danger", "Erro ao atualizar status", "Não foi possível confirmar o novo status da invoice.")
		return
	}

	inv.Status = domain.Status(status)

	writeSuccessHeaders(w, "Pagamento solicitado", "A invoice foi enviada para processamento.", false)

	components.InvoiceRow(inv).Render(r.Context(), w)
}

func (h *WebHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	inv, err := h.service.Repo.FindByID(id)
	if err != nil {
		h.logger.Error("error getting invoice status", map[string]any{"invoice-id": id, "error": err.Error()})
	}

	components.InvoiceRow(inv).Render(r.Context(), w)
}

func writeSuccessHeaders(w http.ResponseWriter, title, message string, resetForm bool) {
	payload := feedbackPayload{
		OperationFeedback: operationFeedback{
			Level:   "success",
			Title:   title,
			Message: message,
		},
		RefreshDashboard: true,
		UpdateCharts:     true,
	}
	if resetForm {
		payload.ResetCreateForm = true
	}

	trigger, _ := json.Marshal(payload)
	w.Header().Set("HX-Trigger", string(trigger))
}

func (h *WebHandler) HandleLogs(w http.ResponseWriter, r *http.Request) {
	if h.logStream == nil {
		http.Error(w, "log stream unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	for _, entry := range h.logStream.Snapshot() {
		fmt.Fprintf(w, "event: message\ndata: %s\n\n", renderLogEntry(entry))
		flusher.Flush()
	}

	entries, cancel := h.logStream.Subscribe(16)
	defer cancel()

	for {
		select {
		case <-r.Context().Done():
			return
		case entry, ok := <-entries:
			if !ok {
				return
			}
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", renderLogEntry(entry))
			flusher.Flush()
		}
	}
}

func (h *WebHandler) HandleDetails(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	inv, err := h.service.Repo.FindByID(id)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusNotFound)
		return
	}

	payments, err := h.payments.FindByInvoiceID(id)
	if err != nil {
		http.Error(w, "Erro ao carregar pagamentos", http.StatusInternalServerError)
		return
	}

	components.ItemDetailsModal(components.InvoiceDetailsData{
		Invoice:  inv,
		Payments: components.SortPaymentsByAttempt(payments),
	}).Render(r.Context(), w)
}

func (h *WebHandler) HandleMetricsUpdate(w http.ResponseWriter, r *http.Request) {
	data := components.ChartData{
		Labels: []string{"pending", "succeeded", "failed"},
		Values: []uint64{
			// Exemplo de como calcular ou buscar os valores atuais
			h.metrics.Pending(),
			h.metrics.Succeeded(),
			h.metrics.Failed(),
		},
	}

	jsonBytes, _ := json.Marshal(data)
	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"updateGraph": %s}`, string(jsonBytes)))
	w.WriteHeader(http.StatusNoContent)
}

func renderLogEntry(entry logging.Entry) string {
	fields := make([]string, 0, len(entry.Fields))
	for key, value := range entry.Fields {
		fields = append(fields, fmt.Sprintf("%s=%v", key, value))
	}
	slices.Sort(fields)

	levelClass := "text-info"
	switch strings.ToUpper(entry.Level) {
	case "ERROR":
		levelClass = "text-danger"
	case "INFO":
		levelClass = "text-success"
	}

	extra := ""
	if len(fields) > 0 {
		extra = fmt.Sprintf("<div class='log-fields text-secondary'>%s</div>", template.HTMLEscapeString(strings.Join(fields, " • ")))
	}

	return fmt.Sprintf(
		"<div class='log-entry'><div class='d-flex justify-content-between gap-2'><span class='%s fw-semibold'>%s</span><span class='text-secondary small'>%s</span></div><div>%s</div>%s</div>",
		levelClass,
		template.HTMLEscapeString(entry.Level),
		template.HTMLEscapeString(entry.Time),
		template.HTMLEscapeString(entry.Message),
		extra,
	)
}
