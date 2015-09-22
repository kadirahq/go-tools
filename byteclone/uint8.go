package byteclone

import (
	"reflect"
	"unsafe"
)

const (
	SzUint8 = 1
)

// EncodeUint8 updates the byte slice to match value
func EncodeUint8(d []byte, v *uint8) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	value := (*uint8)(unsafe.Pointer(head.Data))
	*value = *v
}

// DecodeUint8 updates the value to match the byte slice
func DecodeUint8(d []byte, v *uint8) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	value := (*uint8)(unsafe.Pointer(head.Data))
	*v = *value
}

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

// Read updates the struct to use provided byte slice
// This can be used when it's required to read data from
func (v *Uint8) Read(d []byte) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	v.Value = (*uint8)(unsafe.Pointer(head.Data))
	v.Bytes = d[:SzUint8]
}
