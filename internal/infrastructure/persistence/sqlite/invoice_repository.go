package sqlite

import (
	"database/sql"

	"github.com/rcarvalho-pb/payment_project-go/internal/model/invoice"
)

type InvoiceRepository struct {
	db *sql.DB
}

func (r *InvoiceRepository) Save(inv *invoice.Invoice) error {
	panic("not implemented") // TODO: Implement
}

func (r *InvoiceRepository) FindByID(id string) (*invoice.Invoice, error) {
	panic("not implemented") // TODO: Implement
}

func (r *InvoiceRepository) UpdateStatus(id string, status invoice.Status) error {
	panic("not implemented") // TODO: Implement
}
