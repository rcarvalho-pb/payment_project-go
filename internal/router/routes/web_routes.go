package routes

import (
	"net/http"

	web_handler "github.com/rcarvalho-pb/payment_project-go/internal/handler/web"
)

func getWebRoutes(webHandler *web_handler.WebHandler) []Route {
	const resource = "/invoices"
	return []Route{
		{
			URI:      resource,
			Method:   http.MethodPost,
			Function: webHandler.HandleNewInvoice,
		},
		{
			URI:      resource + "/{id}/status",
			Method:   http.MethodGet,
			Function: webHandler.HandleStatus,
		},
		{
			URI:      resource + "/{id}/pay",
			Method:   http.MethodPost,
			Function: webHandler.HandlePayment,
		},
		{
			URI:      resource + "/{id}/details",
			Method:   http.MethodGet,
			Function: webHandler.HandleDetails,
		},
		{
			URI:      resource + "/metrics/update",
			Method:   http.MethodGet,
			Function: webHandler.HandleMetricsUpdate,
		},
		{
			URI:      "/logs",
			Method:   http.MethodGet,
			Function: webHandler.HandleLogs,
		},
		{
			URI:      "/dashboard/overview",
			Method:   http.MethodGet,
			Function: webHandler.HandleDashboardOverview,
		},
		{
			URI:      "/",
			Method:   http.MethodGet,
			Function: webHandler.HandleIndex,
		},
	}
}
