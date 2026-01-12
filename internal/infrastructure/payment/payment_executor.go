package payment

import "math/rand/v2"

type PaymentExecutor struct{}

func (e *PaymentExecutor) Execute() bool {
	return rand.IntN(100) > 70
}
