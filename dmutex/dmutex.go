package dmutex

// TODO pick a suitable package name
//      we're throttling fn calls here

import "sync"

// Mutex is a double mutex whic is used to block a set of goroutines
type Mutex struct {
	p sync.RWMutex
	s sync.RWMutex
}

// New creates a new double mutex
func New() (m *Mutex) {
	m = &Mutex{}
	m.p.Lock()
	return m
}

// Wait blocks calling goroutine until the next flush
func (m *Mutex) Wait() {
	m.s.RLock()
	m.s.RUnlock()
	m.p.RLock()
	m.p.RUnlock()
}

// Flush releases currently waiting goroutines. It ensures that only goroutines
// which were waiting before running the beforeFlush function gets released.
func (m *Mutex) Flush(beforeFlush func()) {
	m.s.Lock()
	beforeFlush()
	m.p.Unlock()
	m.p.Lock()
	m.s.Unlock()
}
