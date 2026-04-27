package worker

import (
	"context"
	"database/sql"

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
	ScheduleRetry(context.Context, *event.PaymentRequestPayload) error
}

type Recorder interface {
	Record(ctx context.Context, evt *event.Event) error
}

type TransactionalRecorder interface {
	Record(ctx context.Context, evt *event.Event) error
	RecordTx(ctx context.Context, tx *sql.Tx, evt *event.Event) error
}
