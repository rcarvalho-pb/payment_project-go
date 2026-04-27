package metrics

import "sync/atomic"

type OutboxCounters struct {
	recorded      uint64
	published     uint64
	publishFailed uint64
}

func (m *OutboxCounters) IncRecorded() {
	atomic.AddUint64(&m.recorded, 1)
}

func (m *OutboxCounters) IncPublished() {
	atomic.AddUint64(&m.published, 1)
}

func (m *OutboxCounters) IncPublishFailed() {
	atomic.AddUint64(&m.publishFailed, 1)
}

func (m *OutboxCounters) Recorded() uint64 {
	return atomic.LoadUint64(&m.recorded)
}

func (m *OutboxCounters) Published() uint64 {
	return atomic.LoadUint64(&m.published)
}

func (m *OutboxCounters) PublishFailed() uint64 {
	return atomic.LoadUint64(&m.publishFailed)
}
