package outbox

import (
	"context"
	"time"

	"github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
)

type OutboxEvent struct {
	ID            string     `db:"id"`
	CorrelationID string     `db:"correlation_id"`
	Type          event.Type `db:"event_type"`
	Payload       []byte     `db:"payload"`
	Published     bool       `db:"published"`
	CreatedAt     time.Time  `db:"created_at"`
}

type OutboxRepository interface {
	Save(evt *OutboxEvent) error
	FindUnpublished(limit int) ([]*OutboxEvent, []string, error)
	MarkPublished(id []string) error
	CountPending(context.Context) (int, error)
}
