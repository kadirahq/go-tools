package fatomic

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

func BenchmarkWithMutex(b *testing.B) {
	var f float64
	var m sync.Mutex

	b.SetParallelism(1000)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Lock()
			f += 0.1
			m.Unlock()
		}
	})
}

func BenchmarkWithAtomic(b *testing.B) {
	var f float64

	b.SetParallelism(1000)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			AddFloat64(&f, 0.1)
		}
	})
}
