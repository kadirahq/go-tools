package byteclone

import (
	"reflect"
	"unsafe"
)

const (
	szfloat32 = 4
)

// Float32 has a float32 value and a byte slice using the same memory location.
// Any changes done to one of these fields will reflect on the other.
type Float32 struct {
	Value *float32
	Bytes []byte
}

// NewFloat32 will create a new Float32 struct with given byte slice.
// If the slice is nil, a new byte slice will be created for storage.
// If the slice length is less than required length, it will panic.
func NewFloat32(d []byte) *Float32 {
	if d == nil {
		d = make([]byte, szfloat32)
	}

	v := &Float32{}
	v.Read(d[:szfloat32])
	return v
}

func (v *Float32) Read(d []byte) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	v.Value = (*float32)(unsafe.Pointer(head.Data))
	v.Bytes = d[:szfloat32]
}
