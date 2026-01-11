package sqlite

import (
	"github.com/jmoiron/sqlx"
	"github.com/rcarvalho-pb/payment_project-go/internal/infrastructure/outbox"
)

type OutboxRepository struct {
	db *sqlx.DB
}

func NewOutboxRepository(db *sqlx.DB) *OutboxRepository {
	return &OutboxRepository{db}
}

func (r *OutboxRepository) Save(evt *outbox.OutboxEvent) error {
	stmt := `
	INSERT INTO outbox_events (id, event_type, payload, published, created_at) VALUES (?, ?, ?, ?, ?)
	`

	if _, err := r.db.Exec(stmt, evt.ID, evt.Type, evt.Payload, evt.Published, evt.CreatedAt); err != nil {
		return err
	}

	return nil
}

func (r *OutboxRepository) FindUnpublished(limit int) ([]*outbox.OutboxEvent, error) {
	query := `
	SELECT * FROM outbox_events WHERE published = 0 ORDER BY created_at LIMIT ?
	`
	var evts []*outbox.OutboxEvent

	if err := r.db.Select(&evts, query, limit); err != nil {
		return nil, err
	}

	return evts, nil
}

func (r *OutboxRepository) MarkPublished(id string) error {
	stmt := `
	UPDATE outbox_events SET published = 1 WHERE id = ?
	`

	if _, err := r.db.Exec(stmt, id); err != nil {
		return err
	}

	return nil
}
