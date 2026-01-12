package event

import "slices"

type Type uint8

const (
	PaymentRequested Type = iota + 1
	PaymentSucceeded
	PaymentFailed
)

var types = []string{
	"REQUESTED",
	"SUCCEEDED",
	"FAILED",
}

func ToType(s string) uint8 {
	return uint8(slices.Index(types, s))
}

func (t Type) String() string {
	return types[t-1]
}

type Event struct {
	Type    Type
	Payload any
}

type EventPublisher interface {
	Publish(*Event) error
}
