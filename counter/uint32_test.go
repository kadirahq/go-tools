package counter

import "testing"

func TestUint32(t *testing.T) {
	c := &Uint32{}

	for i := uint32(0); i < 5; i++ {
		if i != c.Next() {
			t.Fatal("wrong value")
		}
	}
}
