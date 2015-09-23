package atomicplus

import (
	"sync/atomic"
	"unsafe"
)

// AddFloat64 atomically adds a float64 to a float64 and returns new value
func AddFloat64(addr *float64, delta float64) (new float64) {

	convertedAddr := (*int64)(unsafe.Pointer(addr))

	for {
		old := *addr
		new = old + delta

		convertedOld := *(*int64)(unsafe.Pointer(&old))
		convertedNew := *(*int64)(unsafe.Pointer(&new))

		if atomic.CompareAndSwapInt64(convertedAddr, convertedOld, convertedNew) {
			break
		}
	}

	return
}
