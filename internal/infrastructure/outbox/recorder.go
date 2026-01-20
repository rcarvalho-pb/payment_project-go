package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/observability"
)

type Recorder struct {
	Repo OutboxRepository
}

func (r *Recorder) Record(ctx context.Context, evt *event.Event) error {
	payload, err := json.Marshal(evt.Payload)
	if err != nil {
		return err
	}

	cid, _ := observability.CorrelationIDFromContext(ctx)

	outboxEvt := &OutboxEvent{
		ID:            uuid.NewString(),
		CorrelationID: cid,
		Type:          evt.Type,
		Payload:       payload,
		Published:     false,
		CreatedAt:     time.Now(),
	}

	return r.Repo.Save(outboxEvt)
}
