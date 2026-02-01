package health

import "github.com/jmoiron/sqlx"

type SQLChecker struct {
	DB *sqlx.DB
}

func (c *SQLChecker) Name() string {
	return "sqlite"
}

func (c *SQLChecker) Check() error {
	return c.DB.Ping()
}
