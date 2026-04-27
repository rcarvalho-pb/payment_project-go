package routes

import (
	"fmt"
	"net/http"

	rest_handler "github.com/rcarvalho-pb/payment_project-go/internal/handler/rest"
	web_handler "github.com/rcarvalho-pb/payment_project-go/internal/handler/web"
	"github.com/rcarvalho-pb/payment_project-go/internal/middleware"
)

type Route struct {
	URI      string
	Method   string
	Function func(http.ResponseWriter, *http.Request)
}

func ConfigRouter(r *http.ServeMux, webHandler *web_handler.WebHandler, restHandler *rest_handler.RestHandler) {
	var routes []Route

	webRoutes := getWebRoutes(webHandler)
	restRoutes := getRestRoutes(restHandler)

	routes = append(routes, webRoutes...)
	routes = append(routes, restRoutes...)

	for _, route := range routes {
		r.Handle(fmt.Sprintf("%s %s", route.Method, route.URI), middleware.LogginMiddleware(http.HandlerFunc(route.Function)))
	}
}
