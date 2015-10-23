package counter

import (
	"sync/atomic"
)

// Uint32 returns a counter (uint32)
type Uint32 struct {
	n uint32
}

// Next function return the next value
func (c *Uint32) Next() uint32 {
	return atomic.AddUint32(&c.n, 1) - 1
}
