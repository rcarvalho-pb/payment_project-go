package routes

import (
	"net/http"

	web_handler "github.com/rcarvalho-pb/payment_project-go/internal/handler/web"
)

func getWebRoutes(webHandler *web_handler.WebHandler) []Route {
	const resource = "/"
	return []Route{
		{
			URI:      resource,
			Method:   http.MethodGet,
			Function: webHandler.HandleIndex,
		},
		{
			URI:      resource + "invoices",
			Method:   http.MethodPost,
			Function: webHandler.HandleNewInvoice,
		},
		{
			URI:      resource + "invoices/{id}/pay",
			Method:   http.MethodPost,
			Function: webHandler.HandlePayment,
		},
		{
			URI:      resource + "invoices/{id}/details",
			Method:   http.MethodGet,
			Function: webHandler.HandleDetails,
		},
	}
}
