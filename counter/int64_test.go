package counter

import "testing"

func TestInt64(t *testing.T) {
	c := &Int64{}

	for i := int64(0); i < 5; i++ {
		if i != c.Next() {
			t.Fatal("wrong value")
		}
	}
}
