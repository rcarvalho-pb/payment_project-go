package contracts

type OutboxMetrics interface {
	IncRecorded()
	IncPublished()
	IncPublishFailed()
	Recorded() uint64
	Published() uint64
	PublishFailed() uint64
}
