package metrics

import (
	"sync/atomic"

	"github.com/rcarvalho-pb/payment_project-go/internal/application/contracts"
)

var _ contracts.PaymentMetrics = (*Counters)(nil)

type Counters struct {
	PaymentProcessed uint64
	PaymentSucceeded uint64
	PaymentFailed    uint64
}

func (c *Counters) IncProcessed() {
	atomic.AddUint64(&c.PaymentProcessed, 1)
}

func (c *Counters) IncSucceeded() {
	atomic.AddUint64(&c.PaymentSucceeded, 1)
}

func (c *Counters) IncFailed() {
	atomic.AddUint64(&c.PaymentFailed, 1)
}

func (c *Counters) Processed() uint64 {
	return atomic.LoadUint64(&c.PaymentProcessed)
}

func (c *Counters) Succeeded() uint64 {
	return atomic.LoadUint64(&c.PaymentProcessed)
}

func (c *Counters) Failed() uint64 {
	return atomic.LoadUint64(&c.PaymentProcessed)
}
