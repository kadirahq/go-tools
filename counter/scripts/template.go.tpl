package counter

import (
	"sync/atomic"
)

// {{BG}} returns a counter ({{SM}})
type {{BG}} struct {
	n {{SM}}
}

// Next function return the next value
func (c *{{BG}}) Next() {{SM}} {
	return atomic.Add{{BG}}(&c.n, 1) - 1
}
