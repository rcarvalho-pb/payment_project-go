package sqlite

import (
	"context"

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
	INSERT INTO outbox_events (id, correlation_id, event_type, payload, published, created_at) VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(stmt, evt.ID, evt.CorrelationID, evt.Type, evt.Payload, evt.Published, evt.CreatedAt)
	return err
}

func (r *OutboxRepository) FindUnpublished(limit int) ([]*outbox.OutboxEvent, []string, error) {
	query := `
	SELECT * FROM outbox_events WHERE published = 0 ORDER BY created_at LIMIT ?
	`
	var evts []*outbox.OutboxEvent
	var ids []string

	if err := r.db.Select(&evts, query, limit); err != nil {
		return nil, nil, err
	}

	for _, evt := range evts {
		ids = append(ids, evt.ID)
	}

	return evts, ids, nil
}

func (r *OutboxRepository) MarkPublished(ids []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt := `
	UPDATE outbox_events SET published = 1 WHERE id = ?
	`

	for _, id := range ids {
		if _, err := tx.Exec(stmt, id); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *OutboxRepository) CountPending(ctx context.Context) (int, error) {
	stmt := `
	SELECT COUNT(*) FROM outbox_events WHERE event_type = 1 
	`

	var count int

	if err := r.db.QueryRow(stmt).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}
