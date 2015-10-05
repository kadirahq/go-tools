package hybrid

import (
	"reflect"
	"unsafe"
)

const (
	SzUint16 = 2
)

// EncodeUint16 updates the byte slice to match value
func EncodeUint16(d []byte, v *uint16) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	value := (*uint16)(unsafe.Pointer(head.Data))
	*value = *v
}

// DecodeUint16 updates the value to match the byte slice
func DecodeUint16(d []byte, v *uint16) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	value := (*uint16)(unsafe.Pointer(head.Data))
	*v = *value
}

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

// Read updates the struct to use provided byte slice
// This can be used when it's required to read data from
func (v *Uint16) Read(d []byte) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	v.Value = (*uint16)(unsafe.Pointer(head.Data))
	v.Bytes = d[:SzUint16]
}
