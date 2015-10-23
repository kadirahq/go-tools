package counter

import (
	"sync/atomic"
)

// Int32 returns a counter (int32)
type Int32 struct {
	n int32
}

// Next function return the next value
func (c *Int32) Next() int32 {
	return atomic.AddInt32(&c.n, 1) - 1
}
