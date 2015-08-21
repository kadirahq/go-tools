package secure

import "sync/atomic"

// Bool is a thread safe boolean value
// Uses sync/atomic to maintain thread safety
type Bool struct {
	value *int32
}

// NewBool is the constructor.
// A default value can be set.
func NewBool(value bool) *Bool {
	var n int32
	if value {
		n = 1
	}

	return &Bool{&n}
}

// Get is the getter.
func (v *Bool) Get() bool {
	return atomic.LoadInt32(v.value) == 1
}

// Set is the setter.
func (v *Bool) Set(value bool) {
	if value {
		atomic.StoreInt32(v.value, 1)
	} else {
		atomic.StoreInt32(v.value, 0)
	}
}
