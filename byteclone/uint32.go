package byteclone

import (
	"reflect"
	"unsafe"
)

const (
	SzUint32 = 4
)

// Uint32 has a uint32 value and a byte slice using the same memory location.
// Any changes done to one of these fields will reflect on the other.
type Uint32 struct {
	Value *uint32
	Bytes []byte
}

// NewUint32 will create a new Uint32 struct with given byte slice.
// If the slice is nil, a new byte slice will be created for storage.
// If the slice length is less than required length, it will panic.
func NewUint32(d []byte) *Uint32 {
	if d == nil {
		d = make([]byte, SzUint32)
	}

	v := &Uint32{}
	v.Read(d[:SzUint32])
	return v
}

func (v *Uint32) Read(d []byte) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	v.Value = (*uint32)(unsafe.Pointer(head.Data))
	v.Bytes = d[:SzUint32]
}
