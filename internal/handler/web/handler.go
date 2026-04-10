package web_handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/rcarvalho-pb/payment_project-go/internal/application/invoice"
	"github.com/rcarvalho-pb/payment_project-go/internal/views"
)

type WebHandler struct {
	service invoice.Service
}

func NewWebHandler(service invoice.Service) *WebHandler {
	return &WebHandler{service: service}
}

func (h *WebHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	invoices, err := h.service.ListInvoices()
	if err != nil {
		log.Println("error getting invoices")
	}
	views.Layout(invoices).Render(r.Context(), w)
}

func (h *WebHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		log.Println("error id")
		return
	}

	amount, err := strconv.ParseFloat(r.FormValue("amount"), 10)

	h.service.CreateInvoice()
}
