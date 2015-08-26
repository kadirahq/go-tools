package fsutils

import (
	"os"
	"path"
	"testing"
)

func TestEnsureDir(t *testing.T) {
	tmp := os.TempDir()
	dir := path.Join(tmp, "HmduMi7kgbRHCYA3pRep8pvKbTEbDKkDCoCS43yKddiJsuYqo7")

	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); !os.IsNotExist(err) {
		t.Fatal("directory should not exist")
	}

	if err := EnsureDir(dir); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); os.IsNotExist(err) {
		t.Fatal("directory should exist")
	}
}

func TestEnsureFile(t *testing.T) {
	tmp := os.TempDir()
	file := path.Join(tmp, "cMA48AQNnffBgbK3CzwWg6n4MFD96HvziZAoPqE42jR2YsxH5m")

	if err := os.RemoveAll(file); err != nil {
		t.Fatal(err)
	}

	if _, err := os.OpenFile(file, os.O_RDONLY, 0644); !os.IsNotExist(err) {
		t.Fatal("file should not exist")
	}

	if _, err := EnsureFile(file, 1337); err != nil {
		t.Fatal(err)
	}

	if f, err := os.OpenFile(file, os.O_RDONLY, 0644); os.IsNotExist(err) {
		t.Fatal("file should exist")
	} else {
		info, err := f.Stat()
		if err != nil {
			t.Fatal(err)
		}

		if size := info.Size(); size < 1337 {
			t.Fatal("file size should be at least 1337")
		}

		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
