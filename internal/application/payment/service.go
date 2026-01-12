package payment

import (
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/payment"
	"github.com/rcarvalho-pb/payment_project-go/internal/infrastructure/outbox"
)

type Service struct {
	Repo     payment.Repository
	Recorder outbox.Recorder
}
