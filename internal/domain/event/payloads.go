package event

import "time"

type PaymentRequestPayload struct {
	InvoiceID string
	Amount    int64
	Attempt   int
	CreatedAt time.Time
}

type PaymentSucceededPayload struct {
	InvoiceID  string
	PaymentID  string
	FinishedAt time.Time
}

type PaymentFailedPayload struct {
	InvoiceID  string
	PaymentID  string
	Retryable  bool
	Reason     string
	FinishedAt time.Time
}
