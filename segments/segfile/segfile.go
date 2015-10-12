package segfile

import (
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/kadirahq/go-tools/segments"
)

// LoadFiles laods all existing segment files available
// matching provided base path. The base path should contain
// the path to the segment file and the segment file prefix.
// example: "/path/to/segment/files/prefix_"
func LoadFiles(base string, size int64) (segs []*os.File, err error) {
	segs = []*os.File{}

	for i := 0; true; i++ {
		path := base + strconv.Itoa(i)
		file, err := os.OpenFile(path, os.O_RDWR, 0644)
		if err != nil {
			break
		}

		info, err := file.Stat()
		if err != nil {
			return nil, err
		}

		if sz := info.Size(); sz != size {
			err = segments.ErrSegSize
			return nil, err
		}

		segs = append(segs, file)
	}

	return segs, nil
}

// Store is a collection of segment files. Using a set of segment files can
// be faster than using a single growing file. Also, it allocates faster.
type Store struct {
	segs  []*os.File
	segmx *sync.RWMutex
	base  string
	size  int64
	offs  int64
	offmx *sync.Mutex
}

// New creates a collection of segment files on given path
func New(base string, size int64) (s *Store, err error) {
	segs, err := LoadFiles(base, size)
	if err != nil {
		return nil, err
	}

	s = &Store{
		segs:  segs,
		segmx: &sync.RWMutex{},
		base:  base,
		size:  size,
		offmx: &sync.Mutex{},
	}

	return s, nil
}

// Read implements the io.Reader interface
func (s *Store) Read(p []byte) (n int, err error) {
	s.offmx.Lock()
	n, err = s.ReadAt(p, s.offs)
	s.offs += int64(n)
	s.offmx.Unlock()
	return n, err
}

// Write implements the io.Writer interface
func (s *Store) Write(p []byte) (n int, err error) {
	s.offmx.Lock()
	n, err = s.WriteAt(p, s.offs)
	s.offs += int64(n)
	s.offmx.Unlock()
	return n, err
}

// Slice implements the fs.Slicer interface
func (s *Store) Slice(sz int64) (p []byte, err error) {
	s.offmx.Lock()
	p, err = s.SliceAt(sz, s.offs)
	s.offs += int64(len(p))
	s.offmx.Unlock()
	return p, err
}

// Seek implements the io.Seeker interface
func (s *Store) Seek(offset int64, whence int) (off int64, err error) {
	s.offmx.Lock()
	switch whence {
	case 0:
		// from file start
		s.offs = offset
	case 1:
		// from current
		s.offs += offset
	case 2:
		// from file end
		s.segmx.RLock()
		end := int64(len(s.segs)) * s.size
		s.offs = end + offset
		s.segmx.RUnlock()
	}
	off = s.offs
	s.offmx.Unlock()

	return off, nil
}

// ReadAt implements the io.ReaderAt interface
func (s *Store) ReadAt(p []byte, off int64) (n int, err error) {
	sz := int64(len(p))
	toread := p[:]

	fn := func(i, start, end int64) (stop bool, err error) {
		s.segmx.RLock()
		if i >= int64(len(s.segs)) {
			s.segmx.RUnlock()
			return false, io.EOF
		}
		s.segmx.RUnlock()

		var done int64
		req := end - start
		seg := s.segs[i]

		for done < req {
			c, err := seg.ReadAt(toread, start+done)
			if err != nil {
				return false, err
			}

			n += c
			done += int64(c)
			toread = toread[c:]
		}

		return false, nil
	}

	err = segments.Bounds(s.size, off, off+sz, fn)
	return n, err
}

// WriteAt implements the io.WriterAt interface
func (s *Store) WriteAt(p []byte, off int64) (n int, err error) {
	sz := int64(len(p))
	towrite := p[:]

	fn := func(i, start, end int64) (stop bool, err error) {
		if err := s.ensure(i); err != nil {
			return false, err
		}

		var done int64
		req := end - start
		seg := s.segs[i]

		for done < req {
			c, err := seg.WriteAt(towrite, start+done)
			if err != nil {
				return false, err
			}

			n += c
			done += int64(c)
			towrite = towrite[c:]
		}

		return false, nil
	}

	return n, segments.Bounds(s.size, off, off+sz, fn)
}

// SliceAt implements the fs.SlicerAt interface
func (s *Store) SliceAt(sz, off int64) (p []byte, err error) {
	p = make([]byte, sz)
	n, err := s.ReadAt(p, off)
	if err != nil {
		return nil, err
	}

	return p[:n], nil
}

// Ensure makes sure that data upto given offset exists and are valid.
// This will check from current segment length upto given position.
func (s *Store) Ensure(off int64) (err error) {
	n := off / s.size
	if off%s.size != 0 {
		n++
	}

	return s.ensure(n)
}

// Sync implements the fs.Syncer interface
func (s *Store) Sync() (err error) {
	s.segmx.RLock()
	for _, seg := range s.segs {
		if err := seg.Sync(); err != nil {
			s.segmx.RUnlock()
			return err
		}
	}
	s.segmx.RUnlock()

	return nil
}

// Close implements the io.Closer interface
func (s *Store) Close() (err error) {
	s.segmx.RLock()
	for _, seg := range s.segs {
		if err := seg.Close(); err != nil {
			s.segmx.RUnlock()
			return err
		}
	}
	s.segmx.RUnlock()

	return nil
}

// ensure makes sure that segments upto given index exists and are valid.
// This will check from current segment length upto given position.
func (s *Store) ensure(n int64) (err error) {
	num := int(n)

	// fast path
	s.segmx.RLock()
	if num < len(s.segs) {
		s.segmx.RUnlock()
		return nil
	}
	s.segmx.RUnlock()

	// slow path
	s.segmx.Lock()
	defer s.segmx.Unlock()

	available := len(s.segs)
	if num < available {
		return nil
	}

	for i := available; i <= num; i++ {
		path := s.base + strconv.Itoa(i)
		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}

		info, err := file.Stat()
		if err != nil {
			return err
		}

		if sz := info.Size(); sz != s.size {
			err = segments.ErrSegSize
			return err
		}

		s.segs = append(s.segs, file)
	}

	return nil
}
