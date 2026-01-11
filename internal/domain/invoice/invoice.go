package invoice

import (
	"slices"
	"strings"
	"time"
)

type Status uint8

const (
	StatusPending Status = iota + 1
	StatusProcessing
	StatusPaid
	StatusFailed
	StatusCanceled
)

var status = []string{
	"PENDING",
	"PROCESSING",
	"PAID",
	"FAILED",
	"CANCELED",
}

func ToStatus(s string) uint8 {
	return uint8(slices.Index(status, strings.ToUpper(s)))
}

func (s Status) String() string {
	return status[s-1]
}

type Invoice struct {
	ID        string    `db:"id"`
	Amount    int64     `db:"amount"`
	Status    Status    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	DueDate   time.Time `db:"due_date"`
}

func NewInvoice(id string, amount int64) *Invoice {
	return &Invoice{
		ID:        id,
		Amount:    amount,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		DueDate:   time.Now().Add(1 * time.Hour),
	}
}
