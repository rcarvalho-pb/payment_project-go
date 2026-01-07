package invoice

import (
	"slices"
	"strings"
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
	ID     string
	Amount int64
	Status Status
}
