package hybrid

import (
	"reflect"
	"unsafe"
)

const (
	SzFloat64 = 8
)

// EncodeFloat64 updates the byte slice to match value
func EncodeFloat64(d []byte, v *float64) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	value := (*float64)(unsafe.Pointer(head.Data))
	*value = *v
}

// DecodeFloat64 updates the value to match the byte slice
func DecodeFloat64(d []byte, v *float64) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	value := (*float64)(unsafe.Pointer(head.Data))
	*v = *value
}

// Float64 has a float64 value and a byte slice using the same memory location.
// Any changes done to one of these fields will reflect on the other.
type Float64 struct {
	Value *float64
	Bytes []byte
}

// NewFloat64 will create a new Float64 struct with given byte slice.
// If the slice is nil, a new byte slice will be created for storage.
// If the slice length is less than required length, it will panic.
func NewFloat64(d []byte) *Float64 {
	if d == nil {
		d = make([]byte, SzFloat64)
	}

	v := &Float64{}
	v.Read(d[:SzFloat64])
	return v
}

// Read updates the struct to use provided byte slice
// This can be used when it's required to read data from
func (v *Float64) Read(d []byte) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	v.Value = (*float64)(unsafe.Pointer(head.Data))
	v.Bytes = d[:SzFloat64]
}
