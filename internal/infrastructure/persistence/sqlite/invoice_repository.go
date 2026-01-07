package sqlite

import (
	"database/sql"

	"github.com/rcarvalho-pb/payment_project-go/internal/model/invoice"
)

type InvoiceRepository struct {
	db *sql.DB
}

func (r *InvoiceRepository) Save(_ *invoice.Invoice) error {
	panic("not implemented") // TODO: Implement
}

func (r *InvoiceRepository) FindByID(_ string) (*invoice.Invoice, error) {
	panic("not implemented") // TODO: Implement
}

func (r *InvoiceRepository) UpdateStatus(_ string, _ invoice.Status) error {
	panic("not implemented") // TODO: Implement
}
