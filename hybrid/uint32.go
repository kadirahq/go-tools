package hybrid

import (
	"reflect"
	"unsafe"
)

const (
	SzUint32 = 4
)

// EncodeUint32 updates the byte slice to match value
func EncodeUint32(d []byte, v *uint32) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	value := (*uint32)(unsafe.Pointer(head.Data))
	*value = *v
}

// DecodeUint32 updates the value to match the byte slice
func DecodeUint32(d []byte, v *uint32) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	value := (*uint32)(unsafe.Pointer(head.Data))
	*v = *value
}

// Uint32 has a uint32 value and a byte slice using the same memory location.
// Any changes done to one of these fields will reflect on the other.
type Uint32 struct {
	Value *uint32
	Bytes []byte
}

// NewUint32 will create a new Uint32 struct with given byte slice.
// If the slice is nil, a new byte slice will be created for storage.
// If the slice length is less than required length, it will panic.
func NewUint32(d []byte) *Uint32 {
	if d == nil {
		d = make([]byte, SzUint32)
	}

	v := &Uint32{}
	v.Read(d[:SzUint32])
	return v
}

// Read updates the struct to use provided byte slice
// This can be used when it's required to read data from
func (v *Uint32) Read(d []byte) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	v.Value = (*uint32)(unsafe.Pointer(head.Data))
	v.Bytes = d[:SzUint32]
}
