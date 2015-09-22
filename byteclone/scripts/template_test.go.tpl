package byteclone

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func BinaryEncode{{BG}}(v {{SM}}) []byte {
	b := bytes.NewBuffer(nil)
	binary.Write(b, binary.LittleEndian, v)
	return b.Bytes()
}

func Test{{BG}}(t *testing.T) {
	v := New{{BG}}(nil)

	if !bytes.Equal(v.Bytes, BinaryEncode{{BG}}(0)) || *v.Value != 0 {
		t.Fatal("wrong value")
	}

	*v.Value = 5
	if !bytes.Equal(v.Bytes, BinaryEncode{{BG}}(5)) || *v.Value != 5 {
		t.Fatal("wrong value")
	}

	d := BinaryEncode{{BG}}(10)
	d = append(d, 1, 2, 3, 4, 5)
	v.Read(d)

	if !bytes.Equal(v.Bytes, d[:sz{{SM}}]) || *v.Value != 10 {
		t.Fatal("wrong value")
	}
}

func Benchmark{{BG}}Read(b *testing.B) {
	var d = make([]byte, b.N*sz{{SM}})
	var s = New{{BG}}(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Read(d[i*sz{{SM}}:])
	}
}

func Benchmark{{BG}}BinaryRead(b *testing.B) {
	var v {{SM}}
	var d = make([]byte, b.N*sz{{SM}})
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Read(s, binary.LittleEndian, v)
	}
}

func Benchmark{{BG}}Write(b *testing.B) {
	var s = New{{BG}}(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		*s.Value = 0
	}
}

func Benchmark{{BG}}BinaryWrite(b *testing.B) {
	var v {{SM}}
	var d = make([]byte, b.N*sz{{SM}})
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Write(s, binary.LittleEndian, v)
	}
}
