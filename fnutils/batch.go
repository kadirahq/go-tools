package fnutils

import (
	"sync"
)

// Batch wraps a function to run only once when called multiple times.
// While running, all function calls will be collected for the next batch.
// Calls made before running the payload function will be released.
// This can be useful for synchronizing data (ex. running file.Sync).
type Batch struct {
	running   bool
	payload   func()
	primary   sync.RWMutex
	secondary sync.RWMutex
}

// NewBatch creates a new batch instance using a function.
func NewBatch(fn func()) (b *Batch) {
	b = &Batch{payload: fn}
	b.primary.Lock()
	return b
}

// Run blocks calling goroutine until the next flush
func (b *Batch) Run() {
	b.secondary.RLock()
	b.secondary.RUnlock()
	b.primary.RLock()
	b.primary.RUnlock()
}

// Flush releases currently waiting goroutines. It ensures that only goroutines
// which were waiting before running the beforeFlush function gets released.
func (b *Batch) Flush() {
	b.secondary.Lock()
	b.payload()
	b.primary.Unlock()
	b.primary.Lock()
	b.secondary.Unlock()
}
