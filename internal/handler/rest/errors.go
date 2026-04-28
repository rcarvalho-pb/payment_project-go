package rest_handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	appInvoice "github.com/rcarvalho-pb/payment_project-go/internal/application/invoice"
)

type errorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeAPIError(w http.ResponseWriter, err error) {
	status, code, message := mapInvoiceError(err)
	writeJSON(w, status, errorResponse{
		Error: apiError{
			Code:    code,
			Message: message,
		},
	})
}

func mapInvoiceError(err error) (int, string, string) {
	switch {
	case errors.Is(err, appInvoice.ErrInvalidInvoiceID):
		return http.StatusBadRequest, "invalid_invoice_id", "Informe um identificador válido para a invoice."
	case errors.Is(err, appInvoice.ErrInvalidInvoiceAmount):
		return http.StatusBadRequest, "invalid_invoice_amount", "O valor da invoice deve ser maior que zero."
	case errors.Is(err, appInvoice.ErrInvoiceAlreadyExists):
		return http.StatusConflict, "invoice_already_exists", "Já existe uma invoice com esse identificador."
	case errors.Is(err, appInvoice.ErrInvalidInvoiceState):
		return http.StatusConflict, "invalid_invoice_state", "A invoice não está em um estado que permita pagamento."
	case errors.Is(err, appInvoice.ErrInvoiceNotFound), errors.Is(err, sql.ErrNoRows):
		return http.StatusNotFound, "invoice_not_found", "A invoice informada não foi encontrada."
	default:
		return http.StatusInternalServerError, "internal_error", "Não foi possível concluir a operação agora."
	}
}
