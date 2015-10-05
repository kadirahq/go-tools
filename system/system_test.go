package system

import (
	"fmt"
	"testing"
)

func TestPCPU(t *testing.T) {
	for i := 0; i < 5; i++ {
		fmt.Println("PCPU:", PCPU())
	}
}
