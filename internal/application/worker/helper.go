package worker

import (
	"fmt"
	"math/rand/v2"
	"time"
)

type PaymntExecutorSimulation struct {
	ExecuteFn func() bool
}

func (e *PaymntExecutorSimulation) Execute() bool {
	return e.ExecuteFn()
}

func generateIdempotencyKey(invoiceID string) string {
	return fmt.Sprintf("payment:%s", invoiceID)
}

func generatePaymentID() string {
	return fmt.Sprintf("pay_%d", time.Now().UnixNano())
}

func simulatePayment() bool {
	return rand.IntN(100) < 70
}
