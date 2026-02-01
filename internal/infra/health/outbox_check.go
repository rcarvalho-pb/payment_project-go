package health

import (
	"context"

	"github.com/rcarvalho-pb/payment_project-go/internal/infrastructure/outbox"
)

type OutboxCheck struct {
	Repo outbox.OutboxRepository
}

func (c *OutboxCheck) Name() string {
	return "outbox"
}

func (c *OutboxCheck) Check() error {
	_, err := c.Repo.CountPending(context.Background())
	return err
}
