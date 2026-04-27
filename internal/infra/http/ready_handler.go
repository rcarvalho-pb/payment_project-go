package healthhttp

import (
	"net/http"

	"github.com/rcarvalho-pb/payment_project-go/internal/application/health"
)

type ReadyHandler struct {
	Checks []health.Checker
}

func (h *ReadyHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	for i := range h.Checks {
		if err := h.Checks[i].Check(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(h.Checks[i].Name() + " not ready"))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}
