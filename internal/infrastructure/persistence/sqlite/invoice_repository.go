package sqlite

import (
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

// type Invoice struct {
// 	ID        string    `db:"id"`
// 	Amount    int64     `db:"amount"`
// 	Status    Status    `db:"status"`
// 	CreatedAt time.Time `db:"created_at"`
// 	DueDate   time.Time `db:"due_date"`
// }

func (r *InvoiceRepository) Save(inv *invoice.Invoice) error {
	stmt := `INSERT INTO invoices (id, amount, status, created_at, due_date, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(stmt, inv.ID, inv.Amount, inv.Status, inv.CreatedAt, inv.DueDate, time.Now())
	return err
}

func (r *InvoiceRepository) FindByID(id string) (*invoice.Invoice, error) {
	var inv invoice.Invoice
	query := `SELECT * FROM invoices WHERE id = $1`
	err := r.db.Get(&inv, query, id)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *InvoiceRepository) FindAll() ([]*invoice.Invoice, error) {
	var invs []*invoice.Invoice
	query := `SELECT * FROM invoices`
	err := r.db.Select(&invs, query)
	if err != nil {
		return nil, err
	}
	return invs, nil
}

func (r *InvoiceRepository) UpdateStatus(id string, status invoice.Status) error {
	stmt := `UPDATE invoices SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(stmt, status, time.Now(), id)
	return err
}
