package segmmap

import (
	"bytes"
	"os"
	"testing"

	"github.com/kadirahq/go-tools/segments"
)

var (
	tmpdir  = "/tmp/test-segmmap/"
	tmpfile = tmpdir + "seg_"
)

func setup(t *testing.T) func() {
	if err := os.RemoveAll(tmpdir); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(tmpdir, 0777); err != nil {
		t.Fatal(err)
	}

	return func() {
		if err := os.RemoveAll(tmpdir); err != nil {
			t.Fatal(err)
		}
	}
}

// only works for a small number of bytes
func fill(s *Store, n int) (err error) {
	defer s.Seek(0, 0)
	d := make([]byte, n)

	for i := range d {
		d[i] = byte(i)
	}

	for len(d) > 0 {
		// TODO set manually
		c, err := s.Write(d)
		if err != nil {
			return err
		}

		d = d[c:]
	}

	return nil
}

func TestNew(t *testing.T) {
	defer setup(t)()

	for i := 0; i < 3; i++ {
		s, err := New(tmpfile, 10)
		if err != nil {
			t.Fatal(err)
		}

		if len(s.segs) != 1+1 {
			t.Fatal("wrong length")
		}

		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestOpen(t *testing.T) {
	defer setup(t)()

	s, err := New(tmpfile, 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(s.segs) != 1+1 {
		t.Fatal("wrong length")
	}

	if err := s.Ensure(50); err != nil {
		t.Fatal(err)
	}

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	s, err = New(tmpfile, 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(s.segs) != 6+1 {
		t.Fatal("wrong length")
	}

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSync(t *testing.T) {
	defer setup(t)()

	s, err := New(tmpfile, 10)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Sync(); err != nil {
		t.Fatal(err)
	}

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestReader(t *testing.T) {
	defer setup(t)()

	s, err := New(tmpfile, 3)
	if err != nil {
		t.Fatal(err)
	}

	if err := fill(s, 10); err != nil {
		t.Fatal(err)
	}

	e := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	p := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	if n, err := s.Read(p); err != nil {
		t.Fatal(err)
	} else if n != 10 {
		t.Fatal("short read")
	}

	if !bytes.Equal(p, e) {
		t.Fatal("wrong values")
	}

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestWriter(t *testing.T) {
	defer setup(t)()

	s, err := New(tmpfile, 3)
	if err != nil {
		t.Fatal(err)
	}

	e := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	p := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	if n, err := s.Write(e); err != nil {
		t.Fatal(err)
	} else if n != 10 {
		t.Fatal("short write")
	}

	if n, err := s.ReadAt(p, 0); err != nil {
		t.Fatal(err)
	} else if n != 10 {
		t.Fatal("short read")
	}

	if !bytes.Equal(p, e) {
		t.Fatal("wrong values")
	}

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSlicer(t *testing.T) {
	defer setup(t)()

	s, err := New(tmpfile, 3)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Ensure(10); err != nil {
		t.Fatal(err)
	}

	p, err := s.Slice(10)
	if err != nil {
		t.Fatal(err)
	}

	if len(p) != 3 {
		t.Fatal("wrong length")
	}

	e := []byte{0, 1, 2}
	copy(p, e)

	if n, err := s.ReadAt(p, 0); err != nil {
		t.Fatal(err)
	} else if n != 3 {
		t.Fatal("short read")
	}

	if !bytes.Equal(p, e) {
		t.Fatal("wrong values")
	}

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestReaderAt(t *testing.T) {
	defer setup(t)()

	s, err := New(tmpfile, 3)
	if err != nil {
		t.Fatal(err)
	}

	if err := fill(s, 10); err != nil {
		t.Fatal(err)
	}

	e := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	p := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	if n, err := s.ReadAt(p, 0); err != nil {
		t.Fatal(err)
	} else if n != 10 {
		t.Fatal("short read")
	}

	if !bytes.Equal(p, e) {
		t.Fatal("wrong values")
	}

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestWriterAt(t *testing.T) {
	defer setup(t)()

	s, err := New(tmpfile, 3)
	if err != nil {
		t.Fatal(err)
	}

	e := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	p := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	if n, err := s.WriteAt(e, 0); err != nil {
		t.Fatal(err)
	} else if n != 10 {
		t.Fatal("short write")
	}

	if n, err := s.ReadAt(p, 0); err != nil {
		t.Fatal(err)
	} else if n != 10 {
		t.Fatal("short read")
	}

	if !bytes.Equal(p, e) {
		t.Fatal("wrong values")
	}

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSlicerAt(t *testing.T) {
	defer setup(t)()

	s, err := New(tmpfile, 3)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Ensure(10); err != nil {
		t.Fatal(err)
	}

	p, err := s.SliceAt(10, 1)
	if err != nil {
		t.Fatal(err)
	}

	if len(p) != 2 {
		t.Fatal("wrong length")
	}

	e := []byte{0, 1}
	copy(p, e)

	if n, err := s.ReadAt(p, 1); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal("short read")
	}

	if !bytes.Equal(p, e) {
		t.Fatal("wrong values")
	}

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestImpl(t *testing.T) {
	// throws error if it doesn't
	var _ segments.Store = &Store{}
}
