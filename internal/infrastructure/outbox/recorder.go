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
	var pub bool
	if evt.Type == event.PaymentSucceeded {
		pub = true
	}
	payload, err := json.Marshal(evt.Payload)
	if err != nil {
		return err
	}

	outboxEvt := &OutboxEvent{
		ID:        uuid.NewString(),
		Type:      evt.Type,
		Payload:   payload,
		Published: pub,
		CreatedAt: time.Now(),
	}

	return r.Repo.Save(outboxEvt)
}
