package router

import (
	"net/http"

	web_handler "github.com/rcarvalho-pb/payment_project-go/internal/handler/web"
	"github.com/rcarvalho-pb/payment_project-go/internal/router/routes"
)

func NewRouter(webHandler *web_handler.WebHandler) *http.ServeMux {
	r := http.NewServeMux()

	fs := http.FileServer(http.Dir("./public"))
	r.Handle("GET /static/", http.StripPrefix("/static/", fs))

	routes.ConfigRouter(r, webHandler)

	return r
}
