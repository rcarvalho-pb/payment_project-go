package worker

import (
	"context"
	"testing"
	"time"

	"github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/payment"
	"github.com/rcarvalho-pb/payment_project-go/internal/infrastructure/outbox"
)

type fakePaymentRepo struct {
	savedPayments   []*payment.Payment
	updatedStatuses []payment.Status
}

func (f *fakePaymentRepo) SaveIfNotExist(p *payment.Payment) (bool, error) {
	f.savedPayments = append(f.savedPayments, p)
	return true, nil
}

func (f *fakePaymentRepo) FindByIdempotencyKey(string) (*payment.Payment, error) {
	return nil, nil
}

func (f *fakePaymentRepo) FindByInvoiceID(string) ([]*payment.Payment, error) {
	return nil, nil
}

func (f *fakePaymentRepo) FindAll() ([]*payment.Payment, error) {
	return nil, nil
}

func (f *fakePaymentRepo) UpdateStatus(_ string, status payment.Status) error {
	f.updatedStatuses = append(f.updatedStatuses, status)
	return nil
}

type fakePaymentRecorder struct {
	events []*event.Event
}

func (f *fakePaymentRecorder) Record(_ context.Context, evt *event.Event) error {
	f.events = append(f.events, evt)
	return nil
}

type fakePaymentExecutor struct {
	result bool
}

func (f *fakePaymentExecutor) Execute() bool {
	return f.result
}

type fakePaymentMetrics struct {
	pending   int
	processed int
	succeeded int
	failed    int
}

func (f *fakePaymentMetrics) DeincPending() { f.pending-- }
func (f *fakePaymentMetrics) IncPending()   { f.pending++ }
func (f *fakePaymentMetrics) IncProcessed() { f.processed++ }
func (f *fakePaymentMetrics) IncSucceeded() { f.succeeded++ }
func (f *fakePaymentMetrics) IncFailed()    { f.failed++ }
func (f *fakePaymentMetrics) Pending() uint64 {
	return uint64(max(f.pending, 0))
}
func (f *fakePaymentMetrics) Processed() uint64 {
	return uint64(f.processed)
}
func (f *fakePaymentMetrics) Succeeded() uint64 {
	return uint64(f.succeeded)
}
func (f *fakePaymentMetrics) Failed() uint64 {
	return uint64(f.failed)
}

type fakeRetryOutboxRepo struct{}

func (f *fakeRetryOutboxRepo) Save(*outbox.OutboxEvent) error { return nil }

func (f *fakeRetryOutboxRepo) FindUnpublished(int) ([]*outbox.OutboxEvent, []string, error) {
	return nil, nil, nil
}

func (f *fakeRetryOutboxRepo) MarkPublished([]string) error { return nil }

func (f *fakeRetryOutboxRepo) CountPending(context.Context) (int, error) { return 0, nil }

type fakeRetryOutboxMetrics struct{}

func (f *fakeRetryOutboxMetrics) IncRecorded()      {}
func (f *fakeRetryOutboxMetrics) IncPublished()     {}
func (f *fakeRetryOutboxMetrics) IncPublishFailed() {}
func (f *fakeRetryOutboxMetrics) Recorded() uint64  { return 0 }
func (f *fakeRetryOutboxMetrics) Published() uint64 { return 0 }
func (f *fakeRetryOutboxMetrics) PublishFailed() uint64 {
	return 0
}

func TestPaymentProcessorDoesNotCountTemporaryFailureAsFailedMetric(t *testing.T) {
	repo := &fakePaymentRepo{}
	recorder := &fakePaymentRecorder{}
	metrics := &fakePaymentMetrics{pending: 1}
	processor := PaymentProcessor{
		Repo:     repo,
		Recorder: recorder,
		Retry: &RetryScheduler{
			Recorder:  outbox.Recorder{Repo: &fakeRetryOutboxRepo{}, Metrics: &fakeRetryOutboxMetrics{}},
			MaxRetry:  3,
			BaseDelay: time.Hour,
			MaxDelay:  time.Hour,
		},
		PaymentExecutor: &fakePaymentExecutor{result: false},
		Logger:          &fakeLogger{},
		Metrics:         metrics,
	}

	err := processor.Handle(context.Background(), &event.Event{
		Type: event.PaymentRequested,
		Payload: &event.PaymentRequestPayload{
			InvoiceID: "inv-1",
			Attempt:   1,
		},
	})
	if err != nil {
		t.Fatalf("handle temporary failure: %v", err)
	}

	if got := metrics.Failed(); got != 0 {
		t.Fatalf("expected failed metric to stay at 0 on retryable failure, got %d", got)
	}
	if len(repo.updatedStatuses) != 1 || repo.updatedStatuses[0] != payment.StatusTemporaryFailed {
		t.Fatalf("expected temporary failed status to be persisted, got %v", repo.updatedStatuses)
	}
	if len(recorder.events) != 1 {
		t.Fatalf("expected only the failure event to be recorded immediately, got %d events", len(recorder.events))
	}
	failPayload, ok := recorder.events[0].Payload.(*event.PaymentFailedPayload)
	if !ok {
		t.Fatalf("expected payment failed payload, got %T", recorder.events[0].Payload)
	}
	if !failPayload.Retryable {
		t.Fatal("expected failure payload to be retryable")
	}
}

func TestPaymentProcessorCountsOnlyFinalFailureAsFailedMetric(t *testing.T) {
	repo := &fakePaymentRepo{}
	recorder := &fakePaymentRecorder{}
	metrics := &fakePaymentMetrics{pending: 1}
	processor := PaymentProcessor{
		Repo:            repo,
		Recorder:        recorder,
		Retry:           &RetryScheduler{MaxRetry: 3},
		PaymentExecutor: &fakePaymentExecutor{result: false},
		Logger:          &fakeLogger{},
		Metrics:         metrics,
	}

	err := processor.Handle(context.Background(), &event.Event{
		Type: event.PaymentRequested,
		Payload: &event.PaymentRequestPayload{
			InvoiceID: "inv-1",
			Attempt:   3,
		},
	})
	if err != nil {
		t.Fatalf("handle final failure: %v", err)
	}

	if got := metrics.Failed(); got != 1 {
		t.Fatalf("expected failed metric to increment only on final failure, got %d", got)
	}
	if len(repo.updatedStatuses) != 1 || repo.updatedStatuses[0] != payment.StatusFailed {
		t.Fatalf("expected final failed status to be persisted, got %v", repo.updatedStatuses)
	}
	failPayload, ok := recorder.events[0].Payload.(*event.PaymentFailedPayload)
	if !ok {
		t.Fatalf("expected payment failed payload, got %T", recorder.events[0].Payload)
	}
	if failPayload.Retryable {
		t.Fatal("expected final failure payload to be non-retryable")
	}
}
