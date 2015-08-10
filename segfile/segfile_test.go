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
		"defaults": &Options{Directory: dir},
		"prefixed": &Options{Directory: dir, FilePrefix: "test_"},
		"filesize": &Options{Directory: dir, SegmentSize: 5 * 1024 * 1024},
		"ro-files": &Options{Directory: dir, ReadOnly: true},
		"rw-mmaps": &Options{Directory: dir, MemoryMap: true},
		"ro-mmaps": &Options{Directory: dir, MemoryMap: true, ReadOnly: true},
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
		Directory:  o.Directory,
		FilePrefix: o.FilePrefix,
	})

	if err != nil {
		t.Fatal(err)
	}

	err = sf.MemMap()
	if err != nil {
		t.Fatal(err)
	}

	err = sf.MemLock()
	if err != nil {
		t.Fatal(err)
	}

	err = sf.MUnlock()
	if err != nil {
		t.Fatal(err)
	}

	err = sf.MUnMap()
	if err != nil {
		t.Fatal(err)
	}

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
	off := o.SegmentSize - 2
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
