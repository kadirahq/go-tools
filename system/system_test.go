package system

import (
	"testing"
	"time"
)

func TestPCPU(t *testing.T) {
	if PCPU() != 0 {
		t.Fatal("wrong initial value")
	}

	// wait at least 2 second
	time.Sleep(2 * time.Second)

	if PCPU() <= 0 {
		t.Fatal("must have a valid value")
	}
}
