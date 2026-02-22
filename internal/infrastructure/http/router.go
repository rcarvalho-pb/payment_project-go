package httpapi

import "net/http"

func NewRouter(handler *InvoiceHandler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /invoices", handler.CreateInvoice)
	mux.HandleFunc("POST /invoices/{id}/pay", handler.RequestPayment)
	mux.HandleFunc("GET /invoices", handler.GetInvoices)

	return mux
}
