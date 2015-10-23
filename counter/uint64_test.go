package counter

import "testing"

func TestUint64(t *testing.T) {
	c := &Uint64{}

	for i := uint64(0); i < 5; i++ {
		if i != c.Next() {
			t.Fatal("wrong value")
		}
	}
}
