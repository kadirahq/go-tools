package segmmap

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

// Map is a collection of memory maps. Using a set of memory mapped files can
// be faster than using a single memory map file. Also, it allocates faster.
type Map struct {
	Maps []*memmap.Map
	path string
	size int64
	mutx *sync.RWMutex
}

// NewMap creates a collection of memory maps on given path
func NewMap(path string, size int64) (m *Map, err error) {
	if size == 0 {
		return nil, ErrZeroSz
	}

	m = &Map{
		Maps: []*memmap.Map{},
		path: path,
		size: size,
		mutx: &sync.RWMutex{},
	}

	return m, nil
}

// Load loads a segment file into memory.
func (m *Map) Load(id int64) (f *memmap.Map, err error) {
	// fast path: file already exists
	// RLocks costs lower than Locks
	m.mutx.RLock()
	if id < int64(len(m.Maps)) {
		if f = m.Maps[id]; f != nil {
			m.mutx.RUnlock()
			return f, nil
		}
	}
	m.mutx.RUnlock()

	m.mutx.Lock()
	f, err = m.load(id)
	if err != nil {
		m.mutx.Unlock()
		return nil, err
	}
	m.mutx.Unlock()

	return f, nil
}

// LoadAll loads all existing segment files into memory.
func (m *Map) LoadAll() (err error) {
	dir := path.Dir(m.path)
	base := path.Base(m.path)

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
			if _, err := m.load(id); err != nil {
				// TODO file exists at location but cannot load it
				// should this return an error or load other files?
				continue
			}
		}
	}

	return nil
}

// ReadAt reads data from memory maps starting from offset `off`
func (m *Map) ReadAt(p []byte, off int64) (n int, err error) {
	s := int64(len(p))
	p = p[:0]

	ps, err := m.ZReadAt(s, off)
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
func (m *Map) ZReadAt(sz, off int64) (ps [][]byte, err error) {
	nfiles := sz / m.size
	if off%m.size != 0 {
		nfiles++
	}

	ps = make([][]byte, 0, nfiles)
	sf, ef, so, eo := m.bounds(sz, off)

	for i := sf; i <= ef; i++ {
		var fso int64
		var feo = m.size

		if i == sf {
			fso = so
		}

		if i == ef {
			feo = eo
		}

		m.mutx.Lock()
		f, err := m.load(i)
		if err != nil {
			m.mutx.Unlock()
			return nil, err
		}
		m.mutx.Unlock()

		d := f.Data[fso:feo]
		ps = append(ps, d)
	}

	if err != nil {
		return ps, err
	}

	return ps, nil
}

// WriteAt writes data to memory maps starting from offset `off`
func (m *Map) WriteAt(p []byte, off int64) (n int, err error) {
	sz := int64(len(p))
	sf, ef, so, eo := m.bounds(sz, off)

	for i := sf; i <= ef; i++ {
		var fso int64
		var feo = m.size

		if i == sf {
			fso = so
		}

		if i == ef {
			feo = eo
		}

		m.mutx.Lock()
		f, err := m.load(i)
		if err != nil {
			m.mutx.Unlock()
			return n, err
		}
		m.mutx.Unlock()

		ln := int(feo - fso)
		copy(f.Data[fso:feo], p[n:n+ln])
		n += ln
	}

	// check whether the file after last used file exists
	// if not available load in a background goroutine
	m.prealloc(ef + 1)

	return n, nil
}

// Lock locks all loaded memory maps
func (m *Map) Lock() (err error) {
	for _, f := range m.Maps {
		if err := f.Lock(); err != nil {
			return err
		}
	}

	return nil
}

// Close closes all loaded memory maps
func (m *Map) Close() (err error) {
	for _, f := range m.Maps {
		if err := f.Close(); err != nil {
			return err
		}
	}

	return nil
}

// load creates a memory map and adds it to the map.
// make sure the mutex is locked before running this.
func (m *Map) load(id int64) (f *memmap.Map, err error) {
	count := int64(len(m.Maps))

	if id < count {
		if f = m.Maps[id]; f != nil {
			return f, nil
		}
	}

	idstr := strconv.Itoa(int(id))
	f, err = memmap.NewMap(m.path+idstr, m.size)
	if err != nil {
		return nil, err
	}

	// grow the slice
	if id >= count {
		maps := make([]*memmap.Map, id+1)
		copy(maps, m.Maps)
		m.Maps = maps
	}

	m.Maps[id] = f

	return f, nil
}

func (m *Map) bounds(sz, off int64) (sf, ef, so, eo int64) {
	end := off + sz

	sf = off / m.size
	so = off % m.size
	ef = end / m.size
	eo = end % m.size

	if eo == 0 {
		eo = m.size
		ef--
	}

	return sf, ef, so, eo
}

// prealloc allocates a new file in a background go-routine.
// This is extremely similar to `Load` except the background part.
func (m *Map) prealloc(id int64) {
	// fast path: file already exists
	// RLocks costs lower than Locks
	m.mutx.RLock()
	if id < int64(len(m.Maps)) {
		if f := m.Maps[id]; f != nil {
			m.mutx.RUnlock()
			return
		}
	}
	m.mutx.RUnlock()

	m.mutx.Lock()
	if id < int64(len(m.Maps)) {
		if f := m.Maps[id]; f != nil {
			m.mutx.RUnlock()
			return
		}
	}

	go func() {
		if _, err := m.load(id); err != nil {
			// NOTE: failed to pre-allocate file.
			// We can safely ignore this error.
		}

		m.mutx.Unlock()
	}()
}
