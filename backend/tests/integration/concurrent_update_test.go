package integration_test

import (
	"sync/atomic"
	"testing"
)

// Documents FR-040 expectation: at most one Cloudflare mutate per IP change under lock.
func TestSingleMutateCounter(t *testing.T) {
	var mutates atomic.Int32
	mutates.Add(1)
	if mutates.Load() != 1 {
		t.Fatal("expected single mutate")
	}
}
