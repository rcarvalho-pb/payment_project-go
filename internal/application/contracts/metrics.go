package contracts

type PaymentMetrics interface {
	DeincPending()
	IncPending()
	IncProcessed()
	IncSucceeded()
	IncFailed()
	Pending() uint64
	Processed() uint64
	Succeeded() uint64
	Failed() uint64
}
