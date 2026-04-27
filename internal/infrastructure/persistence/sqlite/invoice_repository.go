package sqlite

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/invoice"
)

type InvoiceRepository struct {
	db *sqlx.DB
}

func NewInvoiceRepository(db *sqlx.DB) *InvoiceRepository {
	return &InvoiceRepository{db}
}

func (r *InvoiceRepository) Save(inv *invoice.Invoice) error {
	stmt := `INSERT INTO invoices (id, amount, status, created_at, due_date, updated_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(stmt, inv.ID, inv.Amount, inv.Status, inv.CreatedAt, inv.DueDate, time.Now())
	return err
}

func (r *InvoiceRepository) FindByID(id string) (*invoice.Invoice, error) {
	var inv invoice.Invoice
	query := `SELECT * FROM invoices WHERE id = ?`
	err := r.db.Get(&inv, query, id)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *InvoiceRepository) FindAll() ([]*invoice.Invoice, error) {
	var invs []*invoice.Invoice
	query := `SELECT * FROM invoices ORDER BY updated_at DESC`
	err := r.db.Select(&invs, query)
	if err != nil {
		return nil, err
	}
	return invs, nil
}

func (r *InvoiceRepository) GetStatus(id string) (uint8, error) {
	stmt := `SELECT status FROM invoices where id = ?`
	var status uint8
	if err := r.db.Get(&status, stmt, id); err != nil {
		return 0, err
	}

	return status, nil
}

func (r *InvoiceRepository) UpdateStatus(id string, status invoice.Status) error {
	return updateInvoiceStatus(r.db, id, status)
}

func (r *InvoiceRepository) UpdateStatusTx(tx *sql.Tx, id string, status invoice.Status) error {
	return updateInvoiceStatus(tx, id, status)
}

func updateInvoiceStatus(execer interface {
	Exec(query string, args ...any) (sql.Result, error)
}, id string, status invoice.Status) error {
	stmt := `UPDATE invoices SET status = ?, updated_at = ? WHERE id = ?`
	result, err := execer.Exec(stmt, status, time.Now(), id)
	if err != nil {
		return err
	}

	if affected, _ := result.RowsAffected(); affected != 1 {
		return errors.New("no row affected")
	}

	return nil
}
