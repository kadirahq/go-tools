package fnutils

import (
	"sync"
)

// Group wraps a function to run only once when called multiple times.
// While running, all function calls will be collected for the next batch.
// Calls made before running the payload function will be released.
// This can be useful for synchronizing data (ex. running file.Sync).
type Group struct {
	running   bool
	payload   func()
	primary   sync.RWMutex
	secondary sync.RWMutex
}

// NewGroup creates a new batch instance using a function.
func NewGroup(fn func()) (g *Group) {
	g = &Group{payload: fn}
	g.primary.Lock()
	return g
}

// Run blocks calling goroutine until the next flush
func (g *Group) Run() {
	g.secondary.RLock()
	g.secondary.RUnlock()
	g.primary.RLock()
	g.primary.RUnlock()
}

// Flush releases currently waiting goroutines. It ensures that only goroutines
// which were waiting before running the beforeFlush function gets released.
func (g *Group) Flush() {
	g.secondary.Lock()
	g.payload()
	g.primary.Unlock()
	g.primary.Lock()
	g.secondary.Unlock()
}
