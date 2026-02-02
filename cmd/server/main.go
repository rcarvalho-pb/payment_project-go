package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rcarvalho-pb/payment_project-go/internal/application/health"
	appInvoice "github.com/rcarvalho-pb/payment_project-go/internal/application/invoice"
	appWorker "github.com/rcarvalho-pb/payment_project-go/internal/application/worker"
	domainEvent "github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
	infrahealth "github.com/rcarvalho-pb/payment_project-go/internal/infra/health"
	healthhttp "github.com/rcarvalho-pb/payment_project-go/internal/infra/http"
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
	db := sqlite.NewDB("./db/db.db")
	// db := sqlite.NewDB("../../db/db.db")
	if db == nil {
		logger.Error("couldn't open db. exiting program", nil)
		os.Exit(1)
	}
	defer db.Close()
	outboxMetrics := metrics.OutboxCounters{}
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
		Metrics:      &outboxMetrics,
		Logger:       &logger,
		PollInterval: 1 * time.Second,
		BatchSize:    10,
	}

	ctx := context.Background()
	defer ctx.Done()

	go dispatcher.Run(ctx)

	paymentExecutor := infraPayment.PaymentExecutor{}

	recorder := outbox.Recorder{
		Repo:    outboxRepo,
		Metrics: &outboxMetrics,
	}

	retry := appWorker.RetryScheduler{
		Recorder:  recorder,
		MaxRetry:  3,
		BaseDelay: 1 * time.Second,
		MaxDelay:  30 * time.Second,
	}

	processor := appWorker.PaymentProcessor{
		Repo:            paymentRepo,
		Recorder:        &recorder,
		Retry:           &retry,
		PaymentExecutor: &paymentExecutor,
		Logger:          &logger,
		Metrics:         &metrics,
	}

	bus.Subscribe(domainEvent.PaymentRequested, processor.Handle)

	invoiceService := appInvoice.Service{
		Repo:     invoiceRepo,
		Recorder: &recorder,
	}

	invoiceHandler := httpapi.NewInvoiceHandler(&invoiceService)

	router := httpapi.NewRouter(invoiceHandler)

	router.Handle("/ready", &healthhttp.ReadyHandler{
		Checks: []health.Checker{
			&infrahealth.SQLChecker{DB: db},
			&infrahealth.OutboxCheck{Repo: outboxRepo},
		},
	})

	router.Handle("/metrics", &healthhttp.MetricsHandler{Counters: &metrics, OutboxMetrics: &outboxMetrics})
	router.HandleFunc("/health", healthhttp.HealthHandler)

	logger.Info("starting server on port :8080", nil)
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
