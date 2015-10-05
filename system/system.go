package system

//#include <time.h>
import "C"
import "time"

var (
	// PCPU has cpu usage as a percentage
	// this value is updated every second
	PCPU float64
)

func init() {
	go setPCPU()
}

func setPCPU() {
	for {
		ticks := C.clock()
		time.Sleep(time.Second)
		clock := float64(C.clock()-ticks) / float64(C.CLOCKS_PER_SEC)
		PCPU = 100 * clock
	}
}
