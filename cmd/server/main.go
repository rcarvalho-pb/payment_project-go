package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	appInvoice "github.com/rcarvalho-pb/payment_project-go/internal/application/invoice"
	appWorker "github.com/rcarvalho-pb/payment_project-go/internal/application/worker"
	domainEvent "github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/logging"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/metrics"
	"github.com/rcarvalho-pb/payment_project-go/internal/infrastructure/eventbus"
	httpapi "github.com/rcarvalho-pb/payment_project-go/internal/infrastructure/http"
	"github.com/rcarvalho-pb/payment_project-go/internal/infrastructure/outbox"
	infraPayment "github.com/rcarvalho-pb/payment_project-go/internal/infrastructure/payment"
	"github.com/rcarvalho-pb/payment_project-go/internal/infrastructure/persistence/sqlite"
)

func main() {
	logger := logging.StdoutLogger{}
	logger.Info("starting program...", nil)
	defer logger.Info("ending program...", nil)
	// db := sqlite.NewDB("./db/db.db")
	db := sqlite.NewDB("../../db/db.db")
	if db == nil {
		logger.Error("couldn't open db. exiting program", nil)
		os.Exit(1)
	}
	defer db.Close()
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
		Recorder: &recorder,
	}

	invoiceHandler := httpapi.NewInvoiceHandler(&invoiceService)

	router := httpapi.NewRouter(invoiceHandler)

	logger.Info("starting server on port :8080", nil)
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
