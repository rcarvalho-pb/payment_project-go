package sqlite

import (
	"github.com/jmoiron/sqlx"
)

func pragmas(db *sqlx.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA foreign_keys = ON;",
		"PRAGMA synchronous = NORMAL;",
	}

	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return err
		}
	}

	return nil
}

func runMigrations(db *sqlx.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS invoices (
			id TEXT PRIMARY KEY,
			amount INTEGER NOT NULL,
			status INTEGER NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			due_date DATETIME NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS payments (
			id TEXT PRIMARY KEY,
			invoice_id TEXT NOT NULL,
			attempt INTEGER NOT NULL,
			status INTEGER NOT NULL,
			idempotency_key TEXT NOT NULL UNIQUE,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS outbox_events (
			id TEXT PRIMARY KEY,
			correlation_id TEXT NOT NULL,
			event_type TEXT NOT NULL,
			payload TEXT NOT NULL,
			published INTEGER NOT NULL,
			created_at DATETIME NOT NULL
		);`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}
