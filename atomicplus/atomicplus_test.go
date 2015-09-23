package atomicplus

import (
	"fmt"
	"sync"
	"testing"
)

func TestAddFloat64(t *testing.T) {
	var testFloat float64
	testFloat = 0.0
	testDelta := 12033343.5
	concurrency := 1000
	expected := float64(concurrency)*testDelta + testFloat

	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			AddFloat64(&testFloat, testDelta)
			// testFloat += testDelta
			wg.Done()
		}()
	}

	wg.Wait()

	fmt.Println(testFloat)

	if testFloat != expected {
		t.Fatal("Not added correctly")
	}
}
