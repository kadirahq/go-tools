package hybrid

import (
	"reflect"
	"unsafe"
)

const (
	SzUint64 = 8
)

// EncodeUint64 updates the byte slice to match value
func EncodeUint64(d []byte, v *uint64) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	value := (*uint64)(unsafe.Pointer(head.Data))
	*value = *v
}

// DecodeUint64 updates the value to match the byte slice
func DecodeUint64(d []byte, v *uint64) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	value := (*uint64)(unsafe.Pointer(head.Data))
	*v = *value
}

// Uint64 has a uint64 value and a byte slice using the same memory location.
// Any changes done to one of these fields will reflect on the other.
type Uint64 struct {
	Value *uint64
	Bytes []byte
}

// NewUint64 will create a new Uint64 struct with given byte slice.
// If the slice is nil, a new byte slice will be created for storage.
// If the slice length is less than required length, it will panic.
func NewUint64(d []byte) *Uint64 {
	if d == nil {
		d = make([]byte, SzUint64)
	}

	v := &Uint64{}
	v.Read(d[:SzUint64])
	return v
}

// Read updates the struct to use provided byte slice
// This can be used when it's required to read data from
func (v *Uint64) Read(d []byte) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	v.Value = (*uint64)(unsafe.Pointer(head.Data))
	v.Bytes = d[:SzUint64]
}
