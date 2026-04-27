package routes

import (
	"net/http"

	rest_handler "github.com/rcarvalho-pb/payment_project-go/internal/handler/rest"
)

func getRestRoutes(restHandler *rest_handler.RestHandler) []Route {
	const resource = "/api/invoices"
	return []Route{
		{
			URI:      resource,
			Method:   http.MethodPost,
			Function: restHandler.CreateInvoice,
		},
		{
			URI:      resource + "/{id}/pay",
			Method:   http.MethodPost,
			Function: restHandler.RequestPayment,
		},
		{
			URI:      resource,
			Method:   http.MethodGet,
			Function: restHandler.GetInvoices,
		},
	}
}
