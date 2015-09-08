package hybrid

import (
	"reflect"
	"unsafe"
)

type Uint32 struct {
	Value *uint32
	Bytes []byte
}

func NewUint32(v uint32) *Uint32 {
	data := make([]byte, 4)
	head := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	vptr := (*uint32)(unsafe.Pointer(head.Data))
	*vptr = v

	return &Uint32{vptr, data}
}
