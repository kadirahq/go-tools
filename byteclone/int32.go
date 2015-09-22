package byteclone

import (
	"reflect"
	"unsafe"
)

const (
	SzInt32 = 4
)

// EncodeInt32 updates the byte slice to match value
func EncodeInt32(d []byte, v *int32) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	value := (*int32)(unsafe.Pointer(head.Data))
	*value = *v
}

// DecodeInt32 updates the value to match the byte slice
func DecodeInt32(d []byte, v *int32) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	value := (*int32)(unsafe.Pointer(head.Data))
	*v = *value
}

// Int32 has a int32 value and a byte slice using the same memory location.
// Any changes done to one of these fields will reflect on the other.
type Int32 struct {
	Value *int32
	Bytes []byte
}

// NewInt32 will create a new Int32 struct with given byte slice.
// If the slice is nil, a new byte slice will be created for storage.
// If the slice length is less than required length, it will panic.
func NewInt32(d []byte) *Int32 {
	if d == nil {
		d = make([]byte, SzInt32)
	}

	v := &Int32{}
	v.Read(d[:SzInt32])
	return v
}

// Read updates the struct to use provided byte slice
// This can be used when it's required to read data from
func (v *Int32) Read(d []byte) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	v.Value = (*int32)(unsafe.Pointer(head.Data))
	v.Bytes = d[:SzInt32]
}
