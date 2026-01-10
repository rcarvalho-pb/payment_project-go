package sqlite

import (
	"errors"
	"log"
	"time"

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
	paymnt, err := r.FindByIdempotencyKey(p.IdempotencyKey)
	if paymnt != nil {
		log.Println("find by idempotency key error: " + err.Error())
		return false, err
	}
	stmt := `
	INSERT INTO payments (id, invoice_id, attempt, status, idempotency_key, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)`
	if _, err := r.db.Exec(stmt, p.ID, p.InvoiceID, p.Attempt, p.Status, p.IdempotencyKey, p.CreatedAt, p.UpdatedAt); err != nil {
		return false, err
	}
	return true, nil
}

func (r *PaymentRepository) FindByIdempotencyKey(idempotencyKey string) (*payment.Payment, error) {
	stmt := `SELECT * FROM payments WHERE idempotency_key = ?`
	var paymt payment.Payment
	err := r.db.Get(&paymt, stmt, idempotencyKey)
	if err != nil {
		return nil, err
	}
	return &paymt, nil
}

func (r *PaymentRepository) UpdateStatus(id string, status payment.Status) error {
	stmt := `UPDATE payments SET status = ?, updated_at = ?  WHERE id = ?`
	affectedRow, err := r.db.Exec(stmt, status, time.Now(), id)
	if err != nil {
		return err
	}

	if i, _ := affectedRow.RowsAffected(); i != 1 {
		return errors.New("no row afected")
	}

	return nil
}

// type Payment struct {
// 	ID             string `db:"id"`
// 	InvoiceID      string `db:"invoice_id"`
// 	Attempt        int    `db:"attempt"`
// 	Status         Status `db:"status"`
// 	IdempotencyKey string `db:"idempotency_key"`
// 	Timestamp      int64  `db:"timestamp"`
// }
