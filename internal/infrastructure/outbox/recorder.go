package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rcarvalho-pb/payment_project-go/internal/application/contracts"
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/observability"
)

type Recorder struct {
	Repo    OutboxRepository
	Metrics contracts.OutboxMetrics
}

func (r *Recorder) Record(ctx context.Context, evt *event.Event) error {
	outboxEvt, err := toOutboxEvent(ctx, evt)
	if err != nil {
		return err
	}

	if err := r.Repo.Save(outboxEvt); err != nil {
		return err
	}

	r.Metrics.IncRecorded()

	return nil
}

func (r *Recorder) RecordTx(ctx context.Context, tx *sql.Tx, evt *event.Event) error {
	repo, ok := r.Repo.(interface {
		SaveTx(*sql.Tx, *OutboxEvent) error
	})
	if !ok {
		return errors.New("outbox repository does not support transactions")
	}

	outboxEvt, err := toOutboxEvent(ctx, evt)
	if err != nil {
		return err
	}

	if err := repo.SaveTx(tx, outboxEvt); err != nil {
		return err
	}

	r.Metrics.IncRecorded()

	return nil
}

func toOutboxEvent(ctx context.Context, evt *event.Event) (*OutboxEvent, error) {
	payload, err := json.Marshal(evt.Payload)
	if err != nil {
		return nil, err
	}

	cid, _ := observability.CorrelationIDFromContext(ctx)

	return &OutboxEvent{
		ID:            uuid.NewString(),
		CorrelationID: cid,
		Type:          evt.Type,
		Payload:       payload,
		Published:     false,
		CreatedAt:     time.Now(),
	}, nil
}
