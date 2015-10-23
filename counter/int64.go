package counter

import (
	"sync/atomic"
)

// Int64 returns a counter (int64)
type Int64 struct {
	n int64
}

// Next function return the next value
func (c *Int64) Next() int64 {
	return atomic.AddInt64(&c.n, 1) - 1
}
