package secure

import "sync"

// Bool is a thread safe boolean value
// Uses sync/atomic to maintain thread safety
type Bool struct {
	sync.RWMutex
	Value bool
}

// NewBool is the constructor.
// A default value can be set.
func NewBool(value bool) *Bool {
	return &Bool{Value: value}
}

// Get is the getter.
func (v *Bool) Get() bool {
	v.RLock()
	value := v.Value
	v.RUnlock()

	return value
}

// Set is the setter.
func (v *Bool) Set(value bool) (changed bool) {
	v.Lock()
	changed = v.Value != value
	v.Value = value
	v.Unlock()

	return changed
}
