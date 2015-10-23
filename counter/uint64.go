package counter

import (
	"sync/atomic"
)

// Uint64 returns a counter (uint64)
type Uint64 struct {
	n uint64
}

// Next function return the next value
func (c *Uint64) Next() uint64 {
	return atomic.AddUint64(&c.n, 1) - 1
}
