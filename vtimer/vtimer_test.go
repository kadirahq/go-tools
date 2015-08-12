package vtimer

import "testing"

func TestSet(t *testing.T) {
	Use(Test)
	Set(123)
	if Now() != 123 {
		t.Fatal("test clock should return preset value")
	}
}
