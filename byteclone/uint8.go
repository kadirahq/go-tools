package byteclone

import (
	"reflect"
	"unsafe"
)

const (
	SzUint8 = 1
)

// Uint8 has a uint8 value and a byte slice using the same memory location.
// Any changes done to one of these fields will reflect on the other.
type Uint8 struct {
	Value *uint8
	Bytes []byte
}

// NewUint8 will create a new Uint8 struct with given byte slice.
// If the slice is nil, a new byte slice will be created for storage.
// If the slice length is less than required length, it will panic.
func NewUint8(d []byte) *Uint8 {
	if d == nil {
		d = make([]byte, SzUint8)
	}

	v := &Uint8{}
	v.Read(d[:SzUint8])
	return v
}

func (v *Uint8) Read(d []byte) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	v.Value = (*uint8)(unsafe.Pointer(head.Data))
	v.Bytes = d[:SzUint8]
}
