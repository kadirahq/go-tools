package byteclone

import (
	"reflect"
	"unsafe"
)

const (
	szfloat64 = 8
)

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
		d = make([]byte, szfloat64)
	}

	v := &Float64{}
	v.Read(d[:szfloat64])
	return v
}

func (v *Float64) Read(d []byte) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	v.Value = (*float64)(unsafe.Pointer(head.Data))
	v.Bytes = d[:szfloat64]
}
