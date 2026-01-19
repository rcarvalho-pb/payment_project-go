package contracts

type PaymentMetrics interface {
	IncProcessed()
	IncSucceeded()
	IncFailed()
}
