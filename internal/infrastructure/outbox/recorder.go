package outbox

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
)

type Recorder struct {
	Repo OutboxRepository
}

func (r *Recorder) Record(evt *event.Event) error {
	payload, err := json.Marshal(evt.Payload)
	if err != nil {
		return err
	}

	outboxEvt := &OutboxEvent{
		ID:        uuid.NewString(),
		Type:      evt.Type,
		Payload:   payload,
		Published: false,
		CreatedAt: time.Now(),
	}

	return r.Repo.Save(outboxEvt)
}
