package sqlite

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rcarvalho-pb/payment_project-go/internal/infra/logging"
)

func NewDB(path string) *sql.DB {
	logger := logging.StdoutLogger{}
	count := 0
	for count < 10 {
		db := connectToDB(path)
		if db == nil {
			logger.Error("error opening DB. backing off for 1 sec...", nil)
			count++
			time.Sleep(1 * time.Second)
			continue
		}
		return db
	}
	return nil
}

func connectToDB(path string) *sql.DB {
	db, err := openDB(path)
	if err != nil {
		return nil
	}

	return db
}

func openDB(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return conn, nil
}
