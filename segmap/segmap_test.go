package segmap

import (
	"bytes"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"testing"
)

var (
	tmpdir  = "/tmp/test-segmmap/"
	tmpfile = tmpdir + "file_"
)

func setup(t *testing.T) {
	if err := os.RemoveAll(tmpdir); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(tmpdir, 0777); err != nil {
		t.Fatal(err)
	}
}

func clear(t *testing.T) {
	if err := os.RemoveAll(tmpdir); err != nil {
		t.Fatal(err)
	}
}

func TestNew(t *testing.T) {
	setup(t)
	defer clear(t)

	for i := 0; i < 3; i++ {
		m, err := NewMap(tmpfile, 10)
		if err != nil {
			t.Fatal(err)
		}

		if len(m.Maps) != 0 {
			t.Fatal("wrong length")
		}

		if err := m.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLoad(t *testing.T) {
	setup(t)
	defer clear(t)

	m, err := NewMap(tmpfile, 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(m.Maps) != 0 {
		t.Fatal("wrong length")
	}

	for i := 0; i < 3; i++ {
		istr := strconv.Itoa(i)
		f, err := os.Create(tmpfile + istr)
		if err != nil {
			t.Fatal(err)
		} else {
			f.Close()
		}

		if _, err := m.Load(int64(i)); err != nil {
			t.Fatal(err)
		}

		if len(m.Maps) != i+1 {
			t.Fatal("wrong length")
		}
	}

	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestLoadAll(t *testing.T) {
	setup(t)
	defer clear(t)

	m, err := NewMap(tmpfile, 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(m.Maps) != 0 {
		t.Fatal("wrong length")
	}

	for i := 0; i < 3; i++ {
		istr := strconv.Itoa(i)
		f, err := os.Create(tmpfile + istr)
		if err != nil {
			t.Fatal(err)
		} else {
			f.Close()
		}
	}

	if err := m.LoadAll(); err != nil {
		t.Fatal(err)
	}

	if len(m.Maps) != 3 {
		t.Fatal("wrong length")
	}

	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestBounds(t *testing.T) {
	m := &Map{size: 10}

	tests := [][]int64{
		// fields => 0:sz, 1:off, 2:sf, 3:ef, 4:so, 5:eo
		[]int64{10, 0, 0, 0, 0, 10}, // complete file
		[]int64{3, 5, 0, 0, 5, 8},   // part of a file
		[]int64{10, 5, 0, 1, 5, 5},  // multiple files 1
		[]int64{15, 5, 0, 1, 5, 10}, // multiple files 2
		[]int64{15, 10, 1, 2, 0, 5}, // multiple files 3
	}

	for _, tst := range tests {
		sf, ef, so, eo := m.bounds(tst[0], tst[1])
		out := []int64{tst[0], tst[1], sf, ef, so, eo}
		if !reflect.DeepEqual(tst, out) {
			t.Fatal("incorrect values", tst, out)
		}
	}
}

func TMapRw(t *testing.T, sz, off int64) {
	setup(t)
	defer clear(t)

	m, err := NewMap(tmpfile, 10)
	if err != nil {
		t.Fatal(err)
	}

	pld := make([]byte, sz)
	for i := 0; i < int(sz); i++ {
		pld[i] = byte(i)
	}

	n, err := m.WriteAt(pld, off)
	if err != nil {
		t.Fatal(err)
	} else if n != int(sz) {
		t.Fatal("wrong size")
	}

	out := make([]byte, sz)
	n, err = m.ReadAt(out, off)
	if err != nil {
		t.Fatal(err)
	} else if n != int(sz) {
		t.Fatal("wrong size")
	}

	if !bytes.Equal(pld, out) {
		t.Fatal("wrong values")
	}

	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestMapRW(t *testing.T) {
	TMapRw(t, 10, 0) // one complete file
	TMapRw(t, 20, 0) // two complete files
	TMapRw(t, 5, 0)  // write from start
	TMapRw(t, 5, 2)  // write from middle
	TMapRw(t, 5, 5)  // write upto end
}

func TestPreAlloc(t *testing.T) {
	setup(t)
	defer clear(t)

	m, err := NewMap(tmpfile, 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(m.Maps) != 0 {
		t.Fatal("wrong length")
	}

	pld := make([]byte, 15)
	for i := 0; i < 15; i++ {
		pld[i] = byte(i)
	}

	if n, err := m.WriteAt(pld, 0); err != nil {
		t.Fatal(err)
	} else if n != 15 {
		t.Fatal("wrong size")
	}

	if len(m.Maps) != 2 {
		t.Fatal("wrong length")
	}

	// run pre-alloc
	runtime.Gosched()

	m.mutx.RLock()
	if len(m.Maps) != 3 {
		t.Fatal("no pre-alloc")
	}
	m.mutx.RUnlock()

	if n, err := m.WriteAt(pld, 10); err != nil {
		t.Fatal(err)
	} else if n != 15 {
		t.Fatal("wrong size")
	}

	// run pre-alloc
	runtime.Gosched()

	m.mutx.RLock()
	if len(m.Maps) != 4 {
		t.Fatal("no pre-alloc")
	}
	m.mutx.RUnlock()

	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
}
