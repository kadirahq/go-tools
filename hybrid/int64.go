package hybrid

import (
	"reflect"
	"unsafe"
)

const (
	SzInt64 = 8
)

// EncodeInt64 updates the byte slice to match value
func EncodeInt64(d []byte, v *int64) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	value := (*int64)(unsafe.Pointer(head.Data))
	*value = *v
}

// DecodeInt64 updates the value to match the byte slice
func DecodeInt64(d []byte, v *int64) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	value := (*int64)(unsafe.Pointer(head.Data))
	*v = *value
}

// Int64 has a int64 value and a byte slice using the same memory location.
// Any changes done to one of these fields will reflect on the other.
type Int64 struct {
	Value *int64
	Bytes []byte
}

// NewInt64 will create a new Int64 struct with given byte slice.
// If the slice is nil, a new byte slice will be created for storage.
// If the slice length is less than required length, it will panic.
func NewInt64(d []byte) *Int64 {
	if d == nil {
		d = make([]byte, SzInt64)
	}

	v := &Int64{}
	v.Read(d[:SzInt64])
	return v
}

// Read updates the struct to use provided byte slice
// This can be used when it's required to read data from
func (v *Int64) Read(d []byte) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	v.Value = (*int64)(unsafe.Pointer(head.Data))
	v.Bytes = d[:SzInt64]
}
