package vtimer

import (
	"time"
)

func init() {
	current = Real
}

var (
	// Real is the real clock
	Real = &real{}

	// Test is the test clock
	Test = &test{}

	// clock in use
	current Clock
)

// Clock gives the time
type Clock interface {
	Now() (ts int64)
}

type real struct {
}

func (c *real) Now() (ts int64) {
	return time.Now().UnixNano()
}

type test struct {
	ts int64
}

func (c *test) Now() (ts int64) {
	return c.ts
}

// Now returns current time
func Now() (ts int64) {
	return current.Now()
}

// Use changes clock in use
func Use(c Clock) {
	current = c
}

// Set changes the time for test clocks
func Set(ts int64) {
	Test.ts = ts
}
