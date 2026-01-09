package main

import (
	"log"
	"os"

	"github.com/rcarvalho-pb/payment_project-go/internal/domain/invoice"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/logging"
	"github.com/rcarvalho-pb/payment_project-go/internal/infrastructure/persistence/sqlite"
)

func main() {
	logger := logging.StdoutLogger{}
	logger.Info("starting program...", nil)
	defer logger.Info("ending program...", nil)
	db := sqlite.NewDB("./db/db.db")
	if db == nil {
		logger.Error("couldn't open db. exiting program", nil)
		os.Exit(1)
	}
	logger.Info("DB opened successfully", nil)

	invoiceRepository := sqlite.NewInvoiceRepository(db)
	// paymentRepository := sqlite.NewPaymentRepository(db)

	logger.Info("finding invoice: inv-123", nil)
	inv, err := invoiceRepository.FindByID("inv-456")
	if err != nil {
		logger.Error("error finding by id: "+err.Error(), nil)
	}

	logger.Info("updating status", nil)
	err = invoiceRepository.UpdateStatus("inv-456", invoice.StatusPaid)
	if err != nil {
		logger.Error("error updating status by id: "+err.Error(), nil)
	}

	inv, err = invoiceRepository.FindByID("inv-456")
	if err != nil {
		logger.Error("error finding by id: "+err.Error(), nil)
	}

	log.Printf("\n%+v\n", inv)
}
