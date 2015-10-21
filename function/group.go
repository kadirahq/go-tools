package function

import (
	"sync"
	"sync/atomic"
)

// Group wraps a function to run only once when called multiple times.
// While running, all function calls will be collected for the next batch.
// Calls made before running the payload function will be released.
// This can be useful for synchronizing data (ex. running file.Sync).
type Group struct {
	running   uint32
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
	// This blocks run calls while a flush is in progress
	// These will have to wait until the next flush call
	g.secondary.RLock()
	g.secondary.RUnlock()

	// mark that the task needs to run
	atomic.StoreUint32(&g.running, 1)

	// This will collect run calls until a flush occurs
	// This will be released immediately after the flush
	g.primary.RLock()
	g.primary.RUnlock()
}

// Flush releases currently waiting goroutines. It ensures that only goroutines
// which were waiting before running the beforeFlush function gets released.
func (g *Group) Flush() {
	// Stop collecting run calls for this flush call
	// Others will have to wait until the next flush call
	g.secondary.Lock()

	// Run the payload function only if run is called
	if atomic.CompareAndSwapUint32(&g.running, 1, 0) {
		g.payload()
	}

	// Release goroutines currently waiting with run
	// and immediately lock again to block future runs
	g.primary.Unlock()
	g.primary.Lock()

	// Unlock run calls locked during the flush
	// This will get flushed with next flush call
	g.secondary.Unlock()
}
