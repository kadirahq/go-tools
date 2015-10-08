package segmap

import (
	"errors"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/kadirahq/go-tools/memmap"
)

var (
	// ErrZeroSz is used when the user attempts to create a memory map
	// with an empty file. Use the size parameter to set required size.
	ErrZeroSz = errors.New("cannot create mmap with empty file")
)

// Store is a collection of memory maps. Using a set of memory mapped files can
// be faster than using a single memory map file. Also, it allocates faster.
type Store struct {
	segs []*memmap.Map
	path string
	size int64
	mutx *sync.RWMutex
}

// New creates a collection of memory maps on given path
func New(path string, size int64) (s *Store, err error) {
	if size == 0 {
		return nil, ErrZeroSz
	}

	s = &Store{
		segs: []*memmap.Map{},
		path: path,
		size: size,
		mutx: &sync.RWMutex{},
	}

	return s, nil
}

// Load loads a segment file into memory.
func (s *Store) Load(id int64) (f *memmap.Map, err error) {
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
	p = p[:0]

	ps, err := s.ZReadAt(sz, off)
	if err != nil {
		return 0, err
	}

	for _, r := range ps {
		n += len(r)
		p = append(p, r...)
	}

	return n, nil
}

// ZReadAt reads data from memory maps starting from offset `off`
// ZReadAt returns a slice of slices from the memory map themselves.
// Data gets read without memory copying but it can be unsafe at times.
// Make sure that the memory map remains mapped while using this data.
// For extended use, make a copy of this data or use the `ReadAt` method.
func (s *Store) ZReadAt(sz, off int64) (ps [][]byte, err error) {
	nfiles := sz / s.size
	if off%s.size != 0 {
		nfiles++
	}

	ps = make([][]byte, 0, nfiles)
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
			return nil, err
		}
		s.mutx.Unlock()

		d := f.Data[fso:feo]
		ps = append(ps, d)
	}

	if err != nil {
		return ps, err
	}

	return ps, nil
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
		copy(f.Data[fso:feo], p[n:n+ln])
		n += ln
	}

	// check whether the file after last used file exists
	// if not available load in a background goroutine
	s.prealloc(ef + 1)

	return n, nil
}

// Length returns the number of segment files loaded
func (s *Store) Length() (n int) {
	s.mutx.RLock()
	n = len(s.segs)
	s.mutx.RUnlock()
	return n
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

// Lock locks all loaded memory maps
func (s *Store) Lock() (err error) {
	for _, f := range s.segs {
		if err := f.Lock(); err != nil {
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
func (s *Store) load(id int64) (f *memmap.Map, err error) {
	count := int64(len(s.segs))

	if id < count {
		if f = s.segs[id]; f != nil {
			return f, nil
		}
	}

	idstr := strconv.Itoa(int(id))
	f, err = memmap.NewMap(s.path+idstr, s.size)
	if err != nil {
		return nil, err
	}

	// grow the slice
	if id >= count {
		maps := make([]*memmap.Map, id+1)
		copy(maps, s.segs)
		s.segs = maps
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
