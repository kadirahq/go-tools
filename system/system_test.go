package system

import (
	"fmt"
	"testing"
	"time"
)

func TestPCPU(t *testing.T) {
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)
		fmt.Printf("PCPU: %.3f%%\n", PCPU)
	}
}
