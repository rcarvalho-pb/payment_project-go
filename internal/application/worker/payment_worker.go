package worker

import (
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
)

type PaymentWorker struct {
	Handler PaymentHandler
}

type PaymentHandler interface {
	Handle(*event.Event) error
}

type PaymentExecutor interface {
	Execute() bool
}

type Scheduler interface {
	Schedule(event.PaymentRequestPayload)
}

type Recorder interface {
	Record(evt *event.Event) error
}
