package counter

import "testing"

func Test{{BG}}(t *testing.T) {
	c := &{{BG}}{}

	for i := {{SM}}(0); i < 5; i++ {
		if i != c.Next() {
			t.Fatal("wrong value")
		}
	}
}
