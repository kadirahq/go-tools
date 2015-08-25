package fnutils

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestBatch(t *testing.T) {
	var n int64

	b := NewBatch(func() {
		time.Sleep(100 * time.Millisecond)
		atomic.AddInt64(&n, 1)
	})

	for i := 0; i < 3; i++ {
		go b.Run()
	}

	// wait to make sure batch is waiting
	time.Sleep(time.Second)

	if atomic.LoadInt64(&n) != 0 {
		t.Fatal("n != 0")
	}

	// start second batch calls
	for i := 0; i < 3; i++ {
		go b.Run()
	}

	// flush first batch
	b.Flush()

	if atomic.LoadInt64(&n) != 1 {
		t.Fatal("n != 1")
	}

	// wait to make sure none of second batch is done
	time.Sleep(time.Second)

	if atomic.LoadInt64(&n) != 1 {
		t.Fatal("n != 1")
	}

	// flush second batch
	b.Flush()

	if atomic.LoadInt64(&n) != 2 {
		t.Fatal("n != 2")
	}
}
