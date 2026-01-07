package payment

import (
	"slices"
	"strings"
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
	ID             string
	InvoiceID      string
	Attempt        int
	Status         Status
	IdempotencyKey string
}
