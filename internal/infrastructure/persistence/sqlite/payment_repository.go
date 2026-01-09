package sqlite

import (
	"github.com/jmoiron/sqlx"
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/payment"
)

type PaymentRepository struct {
	db *sqlx.DB
}

func NewPaymentRepository(db *sqlx.DB) *PaymentRepository {
	return &PaymentRepository{db}
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
