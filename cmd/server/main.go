package main

import (
	"os"

	"github.com/rcarvalho-pb/payment_project-go/internal/infra/logging"
	"github.com/rcarvalho-pb/payment_project-go/internal/infrastructure/persistence/sqlite"
)

func main() {
	logger := logging.StdoutLogger{}
	logger.Info("starting program...", nil)
	db := sqlite.NewDB("./db/db.db")
	if db == nil {
		logger.Error("couldn't open db. exiting program", nil)
		os.Exit(1)
	}
	logger.Info("DB opened successfully", nil)
}
