package sqlite

import (
	"database/sql"

	"github.com/rcarvalho-pb/payment_project-go/internal/model/payment"
)

type PaymentRepository struct {
	db *sql.DB
}

func (r *PaymentRepository) SaveIfNotExist(p *payment.Payment) (bool, error) {
	panic("not implemented") // TODO: Implement
}

func (r *PaymentRepository) FindByIdempotencyKey(idempotencyKey string) (*payment.Payment, error) {
	panic("not implemented") // TODO: Implement
}

func (r *PaymentRepository) UpdateStatus(id string, status payment.Status) error {
	panic("not implemented") // TODO: Implement
}
