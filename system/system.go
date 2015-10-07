// Package system exposes a number of useful information about the system.
package system

//#include <time.h>
import "C"
import "time"

var (
	// pcpu has cpu usage as a percentage
	// this value is updated every second
	pcpu float64
)

func init() {
	go setpcpu()
}

func setpcpu() {
	for {
		ticks := C.clock()
		time.Sleep(time.Second)
		clock := float64(C.clock()-ticks) / float64(C.CLOCKS_PER_SEC)
		pcpu = 100 * clock
	}
}

// PCPU returns the total cpu usage as a percentage. The cpu usage is updated in
// the background every second. CPU usage is claculated by counting cpu ticks.
func PCPU() float64 {
	return pcpu
}
