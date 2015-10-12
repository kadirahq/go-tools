package memmap

import (
	"os"
	"reflect"
	"testing"
)

var (
	tmpfile = "/tmp/test-memmap"
)

func TestNewMap(t *testing.T) {
	if err := os.RemoveAll(tmpfile); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		mmap, err := New(tmpfile, 10)
		if err != nil {
			t.Fatal(err)
		}

		if err := mmap.Close(); err != nil {
			t.Fatal(err)
		}
	}

	if err := os.RemoveAll(tmpfile); err != nil {
		t.Fatal(err)
	}
}

func TestMapFile(t *testing.T) {
	if err := os.RemoveAll(tmpfile); err != nil {
		t.Fatal(err)
	}

	file, err := os.OpenFile(tmpfile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		t.Fatal(err)
	}

	defer file.Close()

	if err := file.Truncate(10); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		mmap, err := MapFile(file, 10)
		if err != nil {
			t.Fatal(err)
		}

		if err := mmap.Close(); err != nil {
			t.Fatal(err)
		}
	}

	if err := os.RemoveAll(tmpfile); err != nil {
		t.Fatal(err)
	}
}

func TestLock(t *testing.T) {
	if err := os.RemoveAll(tmpfile); err != nil {
		t.Fatal(err)
	}

	mmap, err := New(tmpfile, 10)
	if err != nil {
		t.Fatal(err)
	}

	if err := mmap.Lock(); err != nil {
		t.Fatal(err)
	}

	if err := mmap.Close(); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(tmpfile); err != nil {
		t.Fatal(err)
	}
}

func TestData(t *testing.T) {
	if err := os.RemoveAll(tmpfile); err != nil {
		t.Fatal(err)
	}

	mmap, err := New(tmpfile, 10)
	if err != nil {
		t.Fatal(err)
	}

	zeroes := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	values := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	if !reflect.DeepEqual(mmap.Data, zeroes) {
		t.Fatal("mmap data should be empty")
	}

	copy(mmap.Data, values)
	if !reflect.DeepEqual(mmap.Data, values) {
		t.Fatal("mmap data should be empty")
	}

	if err := mmap.Close(); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		mmap, err = New(tmpfile, 10)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(mmap.Data, values) {
			t.Fatal("mmap data should be empty")
		}

		if err := mmap.Close(); err != nil {
			t.Fatal(err)
		}
	}

	if err := os.RemoveAll(tmpfile); err != nil {
		t.Fatal(err)
	}
}
