package web_handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	appInvoice "github.com/rcarvalho-pb/payment_project-go/internal/application/invoice"
)

type feedbackPayload struct {
	OperationFeedback operationFeedback `json:"operationFeedback"`
	RefreshDashboard  bool              `json:"refreshDashboard,omitempty"`
	ResetCreateForm   bool              `json:"resetCreateForm,omitempty"`
	UpdateCharts      bool              `json:"update-charts,omitempty"`
}

type operationFeedback struct {
	Level   string `json:"level"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

func writeHTMXFeedback(w http.ResponseWriter, status int, level, title, message string, opts ...func(*feedbackPayload)) {
	payload := feedbackPayload{
		OperationFeedback: operationFeedback{
			Level:   level,
			Title:   title,
			Message: message,
		},
	}
	for _, opt := range opts {
		opt(&payload)
	}

	trigger, _ := json.Marshal(payload)
	w.Header().Set("HX-Trigger", string(trigger))
	w.WriteHeader(status)
}

func withCharts() func(*feedbackPayload) {
	return func(p *feedbackPayload) {
		p.UpdateCharts = true
	}
}

func withResetCreateForm() func(*feedbackPayload) {
	return func(p *feedbackPayload) {
		p.ResetCreateForm = true
	}
}

func mapInvoiceError(err error) (int, string, string) {
	switch {
	case errors.Is(err, appInvoice.ErrInvalidInvoiceID):
		return http.StatusBadRequest, "Identificador inválido", "Informe um identificador para a invoice."
	case errors.Is(err, appInvoice.ErrInvalidInvoiceAmount):
		return http.StatusBadRequest, "Valor inválido", "O valor da invoice deve ser maior que zero."
	case errors.Is(err, appInvoice.ErrInvoiceAlreadyExists):
		return http.StatusConflict, "Invoice já existe", "Já existe uma invoice com esse identificador."
	case errors.Is(err, appInvoice.ErrInvalidInvoiceState):
		return http.StatusConflict, "Operação não permitida", "Somente invoices pendentes podem ser enviadas para pagamento."
	case errors.Is(err, appInvoice.ErrInvoiceNotFound), errors.Is(err, sql.ErrNoRows):
		return http.StatusNotFound, "Invoice não encontrada", "Não foi possível localizar a invoice informada."
	default:
		return http.StatusInternalServerError, "Erro interno", "Não foi possível concluir a operação agora."
	}
}
