package main

import (
	"context"
	"os"
	"time"

	"github.com/google/uuid"
	appInvoice "github.com/rcarvalho-pb/payment_project-go/internal/application/invoice"
	appWorker "github.com/rcarvalho-pb/payment_project-go/internal/application/worker"
	domainEvent "github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/logging"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/metrics"
	"github.com/rcarvalho-pb/payment_project-go/internal/infrastructure/eventbus"
	"github.com/rcarvalho-pb/payment_project-go/internal/infrastructure/outbox"
	infraPayment "github.com/rcarvalho-pb/payment_project-go/internal/infrastructure/payment"
	"github.com/rcarvalho-pb/payment_project-go/internal/infrastructure/persistence/sqlite"
)

func main() {
	logger := logging.StdoutLogger{}
	logger.Info("starting program...", nil)
	defer logger.Info("ending program...", nil)
	db := sqlite.NewDB("./db/db.db")
	defer db.Close()
	if db == nil {
		logger.Error("couldn't open db. exiting program", nil)
		os.Exit(1)
	}
	metrics := metrics.Counters{}
	bus := eventbus.NewInMemoryBus()

	invoiceRepo := sqlite.NewInvoiceRepository(db)
	invoicePaymentHandler := appInvoice.PaymentEventHandler{
		Repo: invoiceRepo,
	}

	paymentRepo := sqlite.NewPaymentRepository(db)

	bus.Subscribe(domainEvent.PaymentSucceeded, invoicePaymentHandler.Handle)
	bus.Subscribe(domainEvent.PaymentFailed, invoicePaymentHandler.Handle)

	outboxRepo := sqlite.NewOutboxRepository(db)
	dispatcher := appWorker.OutboxDispatcher{
		Repo:         outboxRepo,
		EventBus:     bus,
		Logger:       &logger,
		PollInterval: 1 * time.Second,
		BatchSize:    10,
	}

	ctx := context.Background()
	defer ctx.Done()

	go func() { dispatcher.Run(ctx) }()

	paymentExecutor := infraPayment.PaymentExecutor{}

	recorder := outbox.Recorder{
		Repo: outboxRepo,
	}

	processor := appWorker.PaymentProcessor{
		Repo:            paymentRepo,
		Recorder:        &recorder,
		PaymentExecutor: &paymentExecutor,
		Logger:          &logger,
		Metrics:         metrics,
	}

	bus.Subscribe(domainEvent.PaymentRequested, processor.Handle)

	invoiceService := appInvoice.Service{
		Repo:     invoiceRepo,
		EventBus: bus,
	}

	invID := uuid.NewString()
	invoiceService.CreateInvoice(invID, 100)
	invoiceService.RequestPayment(invID)
}
