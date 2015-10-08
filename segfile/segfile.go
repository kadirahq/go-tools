package segfile

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

var (
	// ErrRead is used when a read didn't complete
	ErrRead = errors.New("bytes read != required size")

	// ErrWrite is used when a write didn't complete
	ErrWrite = errors.New("bytes written != payload size")
)

// Store is a collection of segment files. Using a set of segment files can
// be faster than using a single growing file. Also, it allocates faster.
type Store struct {
	segs []*os.File
	path string
	size int64
	mutx *sync.RWMutex
}

// New creates a collection of segment files on given path
func New(path string, size int64) (s *Store, err error) {
	s = &Store{
		segs: []*os.File{},
		path: path,
		size: size,
		mutx: &sync.RWMutex{},
	}

	return s, nil
}

// Load opens a segment file handler.
func (s *Store) Load(id int64) (f *os.File, err error) {
	// fast path: file already exists
	// RLocks costs lower than Locks
	s.mutx.RLock()
	if id < int64(len(s.segs)) {
		if f = s.segs[id]; f != nil {
			s.mutx.RUnlock()
			return f, nil
		}
	}
	s.mutx.RUnlock()

	s.mutx.Lock()
	f, err = s.load(id)
	if err != nil {
		s.mutx.Unlock()
		return nil, err
	}
	s.mutx.Unlock()

	return f, nil
}

// LoadAll loads all existing segment files into memory.
func (s *Store) LoadAll() (err error) {
	dir := path.Dir(s.path)
	base := path.Base(s.path)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if strings.HasPrefix(name, base) {
			idstr := strings.TrimPrefix(name, base)
			i, err := strconv.Atoi(idstr)
			if err != nil {
				continue
			}

			id := int64(i)
			if _, err := s.load(id); err != nil {
				// TODO file exists at location but cannot load it
				// should this return an error or load other files?
				continue
			}
		}
	}

	return nil
}

// ReadAt reads data from memory maps starting from offset `off`
func (s *Store) ReadAt(p []byte, off int64) (n int, err error) {
	sz := int64(len(p))
	sf, ef, so, eo := s.bounds(sz, off)

	for i := sf; i <= ef; i++ {
		var fso int64
		var feo = s.size

		if i == sf {
			fso = so
		}

		if i == ef {
			feo = eo
		}

		s.mutx.Lock()
		f, err := s.load(i)
		if err != nil {
			s.mutx.Unlock()
			return n, err
		}
		s.mutx.Unlock()

		ln := int(feo - fso)
		dst := p[n : n+ln]

		if c, err := f.ReadAt(dst, fso); err != nil {
			return n, err
		} else if c != ln {
			return n, ErrRead
		}

		n += ln
	}

	return n, nil
}

// WriteAt writes data to memory maps starting from offset `off`
func (s *Store) WriteAt(p []byte, off int64) (n int, err error) {
	sz := int64(len(p))
	sf, ef, so, eo := s.bounds(sz, off)

	for i := sf; i <= ef; i++ {
		var fso int64
		var feo = s.size

		if i == sf {
			fso = so
		}

		if i == ef {
			feo = eo
		}

		s.mutx.Lock()
		f, err := s.load(i)
		if err != nil {
			s.mutx.Unlock()
			return n, err
		}
		s.mutx.Unlock()

		ln := int(feo - fso)
		src := p[n : n+ln]

		if c, err := f.WriteAt(src, fso); err != nil {
			return n, err
		} else if c != ln {
			return n, ErrWrite
		}

		n += ln
	}

	// check whether the file after last used file exists
	// if not available load in a background goroutine
	s.prealloc(ef + 1)

	return n, nil
}

// Sync syncs all loaded memory maps
func (s *Store) Sync() (err error) {
	for _, f := range s.segs {
		if err := f.Sync(); err != nil {
			return err
		}
	}

	return nil
}

// Close closes all loaded memory maps
func (s *Store) Close() (err error) {
	for _, f := range s.segs {
		if err := f.Close(); err != nil {
			return err
		}
	}

	return nil
}

// load creates a memory map and adds it to the map.
// make sure the mutex is locked before running this.
func (s *Store) load(id int64) (f *os.File, err error) {
	count := int64(len(s.segs))

	if id < count {
		if f = s.segs[id]; f != nil {
			return f, nil
		}
	}

	idstr := strconv.Itoa(int(id))
	f, err = os.OpenFile(s.path+idstr, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	// grow the slice
	if id >= count {
		segs := make([]*os.File, id+1)
		copy(segs, s.segs)
		s.segs = segs
	}

	s.segs[id] = f

	return f, nil
}

func (s *Store) bounds(sz, off int64) (sf, ef, so, eo int64) {
	end := off + sz

	sf = off / s.size
	so = off % s.size
	ef = end / s.size
	eo = end % s.size

	if eo == 0 {
		eo = s.size
		ef--
	}

	return sf, ef, so, eo
}

// prealloc allocates a new file in a background go-routine.
// This is extremely similar to `Load` except the background part.
func (s *Store) prealloc(id int64) {
	// fast path: file already exists
	// RLocks costs lower than Locks
	s.mutx.RLock()
	if id < int64(len(s.segs)) {
		if f := s.segs[id]; f != nil {
			s.mutx.RUnlock()
			return
		}
	}
	s.mutx.RUnlock()

	s.mutx.Lock()
	if id < int64(len(s.segs)) {
		if f := s.segs[id]; f != nil {
			s.mutx.RUnlock()
			return
		}
	}

	go func() {
		if _, err := s.load(id); err != nil {
			// NOTE: failed to pre-allocate file.
			// We can safely ignore this error.
		}

		s.mutx.Unlock()
	}()
}
