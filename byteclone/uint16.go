package byteclone

import (
	"reflect"
	"unsafe"
)

const (
	SzUint16 = 2
)

// Uint16 has a uint16 value and a byte slice using the same memory location.
// Any changes done to one of these fields will reflect on the other.
type Uint16 struct {
	Value *uint16
	Bytes []byte
}

// NewUint16 will create a new Uint16 struct with given byte slice.
// If the slice is nil, a new byte slice will be created for storage.
// If the slice length is less than required length, it will panic.
func NewUint16(d []byte) *Uint16 {
	if d == nil {
		d = make([]byte, SzUint16)
	}

	v := &Uint16{}
	v.Read(d[:SzUint16])
	return v
}

func (v *Uint16) Read(d []byte) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	v.Value = (*uint16)(unsafe.Pointer(head.Data))
	v.Bytes = d[:SzUint16]
}
