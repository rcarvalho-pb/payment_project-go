package metrics

import "testing"

func TestDeincPendingDoesNotUnderflowAtZero(t *testing.T) {
	var counters Counters

	counters.DeincPending()

	if got := counters.Pending(); got != 0 {
		t.Fatalf("expected pending to stay at 0, got %d", got)
	}
}
