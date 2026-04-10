package routes

import (
	"fmt"
	"net/http"

	web_handler "github.com/rcarvalho-pb/payment_project-go/internal/handler/web"
)

type Route struct {
	URI      string
	Method   string
	Function func(http.ResponseWriter, *http.Request)
}

func ConfigRouter(r *http.ServeMux, webHandler *web_handler.WebHandler) {
	var routes []Route

	webRoutes := getWebRoutes(webHandler)

	routes = append(routes, webRoutes...)

	for _, route := range routes {
		r.HandleFunc(fmt.Sprintf("%s %s", route.Method, route.URI), http.HandlerFunc(route.Function))
	}
}
