package hybrid

import (
	"bytes"
	"testing"
)

func TestUint32(t *testing.T) {
	s := NewUint32(1)

	if !bytes.Equal(s.Bytes, []byte{1, 0, 0, 0}) || *s.Value != 1 {
		t.Fatal("wrong value")
	}

	*s.Value = 2
	if !bytes.Equal(s.Bytes, []byte{2, 0, 0, 0}) || *s.Value != 2 {
		t.Fatal("wrong value")
	}

	*s.Value = 256
	if !bytes.Equal(s.Bytes, []byte{0, 1, 0, 0}) || *s.Value != 256 {
		t.Fatal("wrong value")
	}
}
