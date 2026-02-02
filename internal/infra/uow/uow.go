package uow

import (
	"database/sql"
)

type UnionOfWork interface {
	Begin() (*sql.Tx, error)
}
