package byteclone

import (
	"reflect"
	"unsafe"
)

const (
	szuint64 = 8
)

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
		d = make([]byte, szuint64)
	}

	v := &Uint64{}
	v.Read(d[:szuint64])
	return v
}

func (v *Uint64) Read(d []byte) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	v.Value = (*uint64)(unsafe.Pointer(head.Data))
	v.Bytes = d[:szuint64]
}
