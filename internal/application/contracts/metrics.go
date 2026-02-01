package contracts

type PaymentMetrics interface {
	IncProcessed()
	IncSucceeded()
	IncFailed()
	Processed() uint64
	Succeeded() uint64
	Failed() uint64
}
