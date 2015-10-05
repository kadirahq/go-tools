package hybrid

import (
	"reflect"
	"unsafe"
)

const (
	SzFloat32 = 4
)

// EncodeFloat32 updates the byte slice to match value
func EncodeFloat32(d []byte, v *float32) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	value := (*float32)(unsafe.Pointer(head.Data))
	*value = *v
}

// DecodeFloat32 updates the value to match the byte slice
func DecodeFloat32(d []byte, v *float32) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	value := (*float32)(unsafe.Pointer(head.Data))
	*v = *value
}

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
		d = make([]byte, SzFloat32)
	}

	v := &Float32{}
	v.Read(d[:SzFloat32])
	return v
}

// Read updates the struct to use provided byte slice
// This can be used when it's required to read data from
func (v *Float32) Read(d []byte) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	v.Value = (*float32)(unsafe.Pointer(head.Data))
	v.Bytes = d[:SzFloat32]
}
