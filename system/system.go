package system

//#include <time.h>
import "C"
import "time"

// PCPU returns cpu usage as a percentage
// Based on http://stackoverflow.com/a/31030753
// TODO verify results (seems okay)
// TODO remove the 1 second sleep
func PCPU() float64 {
	startSeconds := time.Now()
	startTicks := C.clock()
	time.Sleep(time.Second)
	clockSeconds := float64(C.clock()-startTicks) / float64(C.CLOCKS_PER_SEC)
	realSeconds := time.Since(startSeconds).Seconds()
	return clockSeconds / realSeconds * 100
}
