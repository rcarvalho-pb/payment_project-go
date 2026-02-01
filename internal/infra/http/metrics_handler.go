package healthhttp

import (
	"fmt"
	"net/http"

	"github.com/rcarvalho-pb/payment_project-go/internal/application/contracts"
)

type MetricsHandler struct {
	Counters      contracts.PaymentMetrics
	OutboxMetrics contracts.OutboxMetrics
}

func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(
		w,
		"payments_processed %d\npayment_succeeded %d\npayment_failed %d\n",
		h.Counters.Processed(),
		h.Counters.Succeeded(),
		h.Counters.Failed(),
	)

	fmt.Fprintf(
		w,
		"outbox_recorded %d\noutbox_published %d\noutbox_publish_failed %d\n",
		h.OutboxMetrics.Recorded(),
		h.OutboxMetrics.Published(),
		h.OutboxMetrics.PublishFailed(),
	)
}
