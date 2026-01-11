package metrics

import "sync/atomic"

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
