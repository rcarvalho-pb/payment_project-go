package sqlite

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type SQLiteUnionOfWork struct {
	DB *sqlx.DB
}

func New(db *sqlx.DB) *SQLiteUnionOfWork {
	return &SQLiteUnionOfWork{
		DB: db,
	}
}

func (u *SQLiteUnionOfWork) Begin() (*sql.Tx, error) {
	tx, err := u.DB.Begin()
	return tx, err
}
