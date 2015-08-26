package mmap

import (
	"os"
	"reflect"
	"testing"

	"github.com/kadirahq/go-tools/fsutils"
	"github.com/kadirahq/go-tools/logger"
)

var (
	fpath = "/tmp/test-mmap-1"
)

func TestNewMMap(t *testing.T) {
	if err := os.RemoveAll(fpath); err != nil {
		logger.Error(err, "delete file")
		t.Fatal(err)
	}

	if _, err := fsutils.EnsureFile(fpath, 10); err != nil {
		logger.Error(err, "create file")
		t.Fatal(err)
	}

	mmap, err := NewMMap(fpath, true)
	if err != nil {
		logger.Error(err, "create mmap")
		t.Fatal(err)
	}

	if err := mmap.Close(); err != nil {
		logger.Error(err, "close mmap")
		t.Fatal(err)
	}

	if err := os.RemoveAll(fpath); err != nil {
		logger.Error(err, "delete file")
		t.Fatal(err)
	}
}

func TestWriteReadMMap(t *testing.T) {
	if err := os.RemoveAll(fpath); err != nil {
		logger.Error(err, "delete file")
		t.Fatal(err)
	}

	if _, err := fsutils.EnsureFile(fpath, 10); err != nil {
		logger.Error(err, "create file")
		t.Fatal(err)
	}

	mmap, err := NewMMap(fpath, true)
	if err != nil {
		logger.Error(err, "create mmap")
		t.Fatal(err)
	}

	zeroes := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	number := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	if !reflect.DeepEqual(mmap.Data, zeroes) {
		t.Fatal("mmap data should be empty")
	}

	copy(mmap.Data, number)
	if !reflect.DeepEqual(mmap.Data, number) {
		t.Fatal("mmap data should be empty")
	}

	if err := mmap.Close(); err != nil {
		logger.Error(err, "close mmap")
		t.Fatal(err)
	}

	mmap, err = NewMMap(fpath, true)
	if err != nil {
		logger.Error(err, "create mmap")
		t.Fatal(err)
	}

	if !reflect.DeepEqual(mmap.Data, number) {
		t.Fatal("mmap data should be empty")
	}

	if err := mmap.Close(); err != nil {
		logger.Error(err, "close mmap")
		t.Fatal(err)
	}

	if err := os.RemoveAll(fpath); err != nil {
		logger.Error(err, "delete file")
		t.Fatal(err)
	}
}

func TestNewFile(t *testing.T) {
	if err := os.RemoveAll(fpath); err != nil {
		logger.Error(err, "delete file")
		t.Fatal(err)
	}

	file, err := NewFile(fpath, 10, true)
	if err != nil {
		logger.Error(err, "create mfile")
		t.Fatal(err)
	}

	if err := file.Close(); err != nil {
		logger.Error(err, "close mfile")
		t.Fatal(err)
	}

	if err := os.Remove(fpath); err != nil {
		logger.Error(err, "delete file")
		t.Fatal(err)
	}
}

func TestWriteReadFile(t *testing.T) {
	if err := os.RemoveAll(fpath); err != nil {
		logger.Error(err, "delete file")
		t.Fatal(err)
	}

	file, err := NewFile(fpath, 10, true)
	if err != nil {
		logger.Error(err, "create mfile")
		t.Fatal(err)
	}

	if file.Size() != 10 {
		t.Fatal("incorrect size")
	}

	if n, err := file.WriteAt([]byte{1}, 19); err != nil {
		logger.Error(err, "writeAt mfile")
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal("bytes written != payload size")
	}

	if sz := file.Size(); sz != 20 {
		t.Fatal("incorrect size", sz)
	}

	if err := file.Close(); err != nil {
		logger.Error(err, "close mfile")
		t.Fatal(err)
	}

	file, err = NewFile(fpath, 10, true)
	if err != nil {
		logger.Error(err, "create mfile")
		t.Fatal(err)
	}

	p := []byte{0}
	if n, err := file.ReadAt(p, 19); err != nil {
		logger.Error(err, "readAt mfile")
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal("bytes read != payload size")
	}

	if err := file.Close(); err != nil {
		logger.Error(err, "close mfile")
		t.Fatal(err)
	}

	if err := os.Remove(fpath); err != nil {
		logger.Error(err, "delete file")
		t.Fatal(err)
	}
}
