package segfile

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"testing"
)

const (
	dir = "/tmp/sfile"
)

var (
	OptionsSet = map[string]*Options{
		"defaults": &Options{Path: dir},
		"prefixed": &Options{Path: dir, Prefix: "test_"},
		"filesize": &Options{Path: dir, FileSize: 5 * 1024 * 1024},
		"ro-files": &Options{Path: dir, ReadOnly: true},
		"rw-mmaps": &Options{Path: dir, MemoryMap: true},
		"ro-mmaps": &Options{Path: dir, MemoryMap: true, ReadOnly: true},
	}
)

func TNewWithOptions(t *testing.T, o *Options) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatal(err)
	}

	sf, err := New(o)
	if err != nil {
		t.Fatal(err)
	}

	// run pre-alloc go routines
	runtime.Gosched()

	err = sf.Close()
	if err != nil {
		t.Fatal(err)
	}

	sf, err = New(&Options{
		Path:   o.Path,
		Prefix: o.Prefix,
	})

	if err != nil {
		t.Fatal(err)
	}

	err = sf.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = os.RemoveAll(dir)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNew(t *testing.T) {
	for k, o := range OptionsSet {
		fmt.Println(" - testing with options:", k)
		TNewWithOptions(t, o)
	}
}

func TWriteAtReadAtWithOptions(t *testing.T, o *Options) {
	ro := o.ReadOnly
	o.ReadOnly = false

	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatal(err)
	}

	sf, err := New(o)
	if err != nil {
		t.Fatal(err)
	}

	p := []byte{1, 2, 3, 4}
	off := o.FileSize - 2
	n, err := sf.WriteAt(p, off)
	if err != nil {
		t.Fatal(err)
	} else if n != 4 {
		t.Fatal("write error")
	}

	err = sf.Close()
	if err != nil {
		t.Fatal(err)
	}

	o.ReadOnly = ro
	sf, err = New(o)
	if err != nil {
		t.Fatal(err)
	}

	r := make([]byte, 4)
	n, err = sf.ReadAt(r, off)
	if err != nil {
		t.Fatal(err)
	} else if n != 4 {
		t.Fatal("read error")
	}

	if !reflect.DeepEqual(p, r) {
		t.Fatal("wrong values")
	}

	err = sf.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = os.RemoveAll(dir)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWriteAtReadAt(t *testing.T) {
	for k, o := range OptionsSet {
		fmt.Println(" - testing with options:", k)
		TWriteAtReadAtWithOptions(t, o)
	}
}

func BAllocateWithSegmentSize(b *testing.B, o *Options) {
	if b.N > 100 {
		b.N = 100
	}

	err := os.RemoveAll(dir)
	if err != nil {
		b.Fatal(err)
	}

	sf, err := New(o)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.SetParallelism(10)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sf.Grow(o.FileSize)
		}
	})
}

func BenchmarkAllocateWithSegmentSize_10M(b *testing.B) {
	o := &Options{Path: dir, FileSize: 10 * 1024 * 1024}
	BAllocateWithSegmentSize(b, o)
}

func BenchmarkAllocateWithSegmentSize_20M(b *testing.B) {
	o := &Options{Path: dir, FileSize: 20 * 1024 * 1024}
	BAllocateWithSegmentSize(b, o)
}

func BenchmarkAllocateWithSegmentSize_100M(b *testing.B) {
	o := &Options{Path: dir, FileSize: 100 * 1024 * 1024}
	BAllocateWithSegmentSize(b, o)
}

func BWriteWithPayloadSize(b *testing.B, o *Options, size int) {
	if b.N > 10000 {
		b.N = 10000
	}

	err := os.RemoveAll(dir)
	if err != nil {
		b.Fatal(err)
	}

	sf, err := New(o)
	if err != nil {
		b.Fatal(err)
	}

	p := make([]byte, size)
	n, err := sf.Write(p)
	if err != nil {
		b.Fatal(err)
	} else if n != size {
		b.Fatal("write error")
	}

	b.ResetTimer()
	b.SetParallelism(10)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sf.Write(p)
		}
	})
}

func BenchmarkFileWriteWithPayloadSize_10B(b *testing.B) {
	o := &Options{Path: dir}
	BWriteWithPayloadSize(b, o, 10)
}

func BenchmarkFileWriteWithPayloadSize_20B(b *testing.B) {
	o := &Options{Path: dir}
	BWriteWithPayloadSize(b, o, 20)
}

func BenchmarkFileWriteWithPayloadSize_100B(b *testing.B) {
	o := &Options{Path: dir}
	BWriteWithPayloadSize(b, o, 100)
}

func BenchmarkMMapWriteWithPayloadSize_10B(b *testing.B) {
	o := &Options{Path: dir, MemoryMap: true}
	BWriteWithPayloadSize(b, o, 10)
}

func BenchmarkMMapWriteWithPayloadSize_20B(b *testing.B) {
	o := &Options{Path: dir, MemoryMap: true}
	BWriteWithPayloadSize(b, o, 20)
}

func BenchmarkMMapWriteWithPayloadSize_100B(b *testing.B) {
	o := &Options{Path: dir, MemoryMap: true}
	BWriteWithPayloadSize(b, o, 100)
}
