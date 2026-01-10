package payment

import (
	"slices"
	"strings"
	"time"
)

type Status uint8

const (
	StatusCreated Status = iota + 1
	StatusProcessing
	StatusSuccess
	StatusFailed
)

var status = []string{
	"CREATED",
	"PROCESSING",
	"SUCCESS",
	"FAILED",
}

func ToStatus(s string) uint8 {
	return uint8(slices.Index(status, strings.ToUpper(s)))
}

func (s Status) String() string {
	return status[s-1]
}

type Payment struct {
	ID             string    `db:"id"`
	InvoiceID      string    `db:"invoice_id"`
	Attempt        int       `db:"attempt"`
	Status         Status    `db:"status"`
	IdempotencyKey string    `db:"idempotency_key"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

func NewPayment(id, invoiceID, idempotencyKey string) *Payment {
	return &Payment{
		ID:             id,
		InvoiceID:      invoiceID,
		Attempt:        1,
		Status:         StatusCreated,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}
