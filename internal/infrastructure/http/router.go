package httpapi

import "net/http"

func NewRouter(handler *InvoiceHandler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handler.Index)
	mux.HandleFunc("GET /invoices", handler.ListInvoices)
	mux.HandleFunc("POST /invoices", handler.CreateInvoiceWeb)
	mux.HandleFunc("POST /invoices/{id}/pay", handler.PayInvoice)
	mux.HandleFunc("GET /invoices/{id}", handler.GetInvoice)
	mux.HandleFunc("GET /invoices/{id}/row", handler.InvoiceRow)

	mux.HandleFunc("POST /api/invoices", handler.CreateInvoice)
	mux.HandleFunc("POST /api/invoices/{id}/pay", handler.RequestPayment)
	mux.HandleFunc("GET /api/invoices", handler.GetInvoices)

	return mux
}
