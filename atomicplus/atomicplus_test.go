package atomicplus

import (
	"sync"
	"testing"
)

func TestAddFloat64(t *testing.T) {
	testFloat := 0.0
	testDelta := 3.141592
	concurrency := 10000000
	expected := testFloat

	// Prepare expected value
	for i := 0; i < concurrency; i++ {
		expected += testDelta
	}

	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			AddFloat64(&testFloat, testDelta)
			wg.Done()
		}()
	}

	wg.Wait()

	if testFloat != expected {
		t.Fatal("Not added correctly. Expected:", expected, "Got:", testFloat)
	}
}
