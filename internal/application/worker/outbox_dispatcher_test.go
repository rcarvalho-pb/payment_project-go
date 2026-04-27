package worker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_project-go/internal/infrastructure/outbox"
)

type fakeOutboxRepo struct {
	events    []*outbox.OutboxEvent
	ids       []string
	markedIDs []string
	findErr   error
	markErr   error
}

func (f *fakeOutboxRepo) Save(evt *outbox.OutboxEvent) error { return nil }

func (f *fakeOutboxRepo) FindUnpublished(limit int) ([]*outbox.OutboxEvent, []string, error) {
	_ = limit
	return f.events, f.ids, f.findErr
}

func (f *fakeOutboxRepo) MarkPublished(ids []string) error {
	f.markedIDs = append([]string{}, ids...)
	return f.markErr
}

func (f *fakeOutboxRepo) CountPending(context.Context) (int, error) { return 0, nil }

type fakePublisher struct{}

func (f *fakePublisher) Publish(ctx context.Context, evt *event.Event) error {
	_ = ctx
	payload, ok := evt.Payload.(*event.PaymentRequestPayload)
	if ok && payload.InvoiceID == "fail" {
		return errors.New("publish failed")
	}
	return nil
}

type fakeOutboxMetrics struct {
	published     int
	publishFailed int
}

func (f *fakeOutboxMetrics) IncRecorded()      {}
func (f *fakeOutboxMetrics) IncPublished()     { f.published++ }
func (f *fakeOutboxMetrics) IncPublishFailed() { f.publishFailed++ }
func (f *fakeOutboxMetrics) Recorded() uint64  { return 0 }
func (f *fakeOutboxMetrics) Published() uint64 { return uint64(f.published) }
func (f *fakeOutboxMetrics) PublishFailed() uint64 {
	return uint64(f.publishFailed)
}

type fakeLogger struct{}

func (f *fakeLogger) Info(string, map[string]any)  {}
func (f *fakeLogger) Error(string, map[string]any) {}

func TestOutboxDispatcherMarksOnlySuccessfullyPublishedEvents(t *testing.T) {
	successPayload, err := json.Marshal(&event.PaymentRequestPayload{
		InvoiceID: "ok",
		Attempt:   1,
	})
	if err != nil {
		t.Fatalf("marshal success payload: %v", err)
	}

	failPayload, err := json.Marshal(&event.PaymentRequestPayload{
		InvoiceID: "fail",
		Attempt:   1,
	})
	if err != nil {
		t.Fatalf("marshal fail payload: %v", err)
	}

	repo := &fakeOutboxRepo{
		events: []*outbox.OutboxEvent{
			{ID: "evt-1", Type: event.PaymentRequested, Payload: successPayload, CreatedAt: time.Now()},
			{ID: "evt-2", Type: event.PaymentRequested, Payload: failPayload, CreatedAt: time.Now()},
		},
		ids: []string{"evt-1", "evt-2"},
	}
	metrics := &fakeOutboxMetrics{}
	dispatcher := OutboxDispatcher{
		Repo:         repo,
		EventBus:     &fakePublisher{},
		Metrics:      metrics,
		Logger:       &fakeLogger{},
		PollInterval: time.Second,
		BatchSize:    10,
	}

	dispatcher.dispatchOnce(context.Background())

	if len(repo.markedIDs) != 1 || repo.markedIDs[0] != "evt-1" {
		t.Fatalf("expected only successful event to be marked published, got %v", repo.markedIDs)
	}
	if metrics.published != 1 {
		t.Fatalf("expected 1 published metric, got %d", metrics.published)
	}
	if metrics.publishFailed != 1 {
		t.Fatalf("expected 1 failed publish metric, got %d", metrics.publishFailed)
	}
}
