package sqlite

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/payment"
)

func TestSaveIfNotExistReturnsFalseForDuplicateIdempotencyKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewPaymentRepository(sqlxDB)

	first := &payment.Payment{
		ID:             "pay-1",
		InvoiceID:      "inv-1",
		Attempt:        1,
		Status:         payment.StatusCreated,
		IdempotencyKey: "inv-1:1",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	second := &payment.Payment{
		ID:             "pay-2",
		InvoiceID:      "inv-1",
		Attempt:        1,
		Status:         payment.StatusCreated,
		IdempotencyKey: "inv-1:1",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	mock.ExpectExec("INSERT OR IGNORE INTO payments").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT OR IGNORE INTO payments").
		WillReturnResult(sqlmock.NewResult(0, 0))

	saved, err := repo.SaveIfNotExist(first)
	if err != nil {
		t.Fatalf("save first payment: %v", err)
	}
	if !saved {
		t.Fatal("expected first payment to be saved")
	}

	saved, err = repo.SaveIfNotExist(second)
	if err != nil {
		t.Fatalf("save duplicate payment: %v", err)
	}
	if saved {
		t.Fatal("expected duplicate payment to be ignored")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
