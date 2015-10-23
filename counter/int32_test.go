package counter

import "testing"

func TestInt32(t *testing.T) {
	c := &Int32{}

	for i := int32(0); i < 5; i++ {
		if i != c.Next() {
			t.Fatal("wrong value")
		}
	}
}
