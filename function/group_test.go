package function

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGroup(t *testing.T) {
	var n int64

	g := NewGroup(func() {
		atomic.AddInt64(&n, 1)
	})

	wg := sync.WaitGroup{}
	wg.Add(1e6)

	for i := 0; i < 1e6; i++ {
		go func() {
			wg.Done()
			g.Run()
		}()
	}

	// wait!
	wg.Wait()

	if atomic.LoadInt64(&n) != 0 {
		t.Fatal("n != 0")
	}

	// start second batch calls
	for i := 0; i < 1e6; i++ {
		go g.Run()
	}

	// first
	g.Flush()

	// check the flush counter
	if atomic.LoadInt64(&n) != 1 {
		t.Fatal("n != 1")
	}

	// check again after a while
	time.Sleep(time.Millisecond)
	if atomic.LoadInt64(&n) != 1 {
		t.Fatal("n != 1")
	}

	// second
	g.Flush()

	if atomic.LoadInt64(&n) != 2 {
		t.Fatal("n != 2")
	}
}
