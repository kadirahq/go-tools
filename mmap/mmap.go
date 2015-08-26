package mmap

import (
	"errors"
	"io"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"

	goerr "github.com/go-errors/errors"
	"github.com/kadirahq/go-tools/fsutils"
	"github.com/kadirahq/go-tools/secure"
)

const (
	flag = syscall.MAP_SHARED
	prot = syscall.PROT_READ | syscall.PROT_WRITE
	msa1 = syscall.MS_SYNC
)

var (
	// ErrClosed is returned when the resource is closed
	ErrClosed = errors.New("cannot use closed resource")
)

// MMap is a struct which abstracts memory map system calls and provides a fast
// and easy to use api. The MMap should be unmapped when not in use.
type MMap struct {
	Data []byte
	file *os.File
	lock bool
	hlen uintptr
	hadr uintptr
	open *secure.Bool
}

// NewMMap creates a new memory map struct. This struct contains map data and
// other information needed to manipulate the memory map later.
func NewMMap(path string, lock bool) (m *MMap, err error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, goerr.Wrap(err, 0)
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, goerr.Wrap(err, 0)
	}

	fd := file.Fd()
	sz := info.Size()

	data, err := syscall.Mmap(int(fd), 0, int(sz), prot, flag)
	if err != nil {
		file.Close()
		return nil, goerr.Wrap(err, 0)
	}

	if lock {
		if err := syscall.Mlock(data); err != nil {
			// TODO handle failure to lock memory
		}
	}

	// get slice header to get memory address and length
	head := (*reflect.SliceHeader)(unsafe.Pointer(&data))

	m = &MMap{
		Data: data,
		file: file,
		lock: lock,
		hadr: head.Data,
		hlen: uintptr(head.Len),
		open: secure.NewBool(true),
	}

	return m, nil
}

// Sync synchronizes the memory map with the mapped file. This can be used to
// ensure that all data is written to the disk successfully. Calling the Sync
// method is necessary to survive OS kernel level panics and crashes.
func (m *MMap) Sync() (err error) {
	if !m.open.Get() {
		return goerr.Wrap(ErrClosed, 0)
	}

	_, _, errno := syscall.Syscall(syscall.SYS_MSYNC, m.hadr, m.hlen, msa1)
	if errno != 0 {
		err := syscall.Errno(errno)
		return goerr.Wrap(err, 0)
	}

	return nil
}

// Close unmaps data and closes the file handler. Changes done to the memory
// map will be synced to the disk before closing to prevent data loss.
func (m *MMap) Close() (err error) {
	if !m.open.Get() {
		return goerr.Wrap(ErrClosed, 0)
	}

	if err := m.Sync(); err != nil {
		return goerr.Wrap(err, 0)
	}

	if m.lock {
		if err := syscall.Munlock(m.Data); err != nil {
			return goerr.Wrap(err, 0)
		}
	}

	if err := syscall.Munmap(m.Data); err != nil {
		return goerr.Wrap(err, 0)
	}

	m.open.Set(false)
	return nil
}

// File is similary to os.File but underneath it uses memory maps in order to
// speed up reads and writes. File type also implements many io interfaces
// such as io.Reader, io.ReaderAt, io.Writer, io.WriterAt, io.Closer, io.Seeker
type File struct {
	file *os.File
	data *MMap
	path string
	size int64
	offs int64
	lock bool
	open *secure.Bool
	rwmx sync.RWMutex
	iomx sync.RWMutex
}

// NewFile creates a new memory mapped file handler using the file on path.
// If a size is given, it ensures that the file size is at least that value.
func NewFile(path string, sz int64, lock bool) (f *File, err error) {
	size, err := fsutils.EnsureFile(path, sz)
	if err != nil {
		return nil, goerr.Wrap(err, 0)
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, goerr.Wrap(err, 0)
	}

	data, err := NewMMap(path, lock)
	if err != nil {
		return nil, goerr.Wrap(err, 0)
	}

	f = &File{
		file: file,
		data: data,
		path: path,
		size: size,
		lock: lock,
		open: secure.NewBool(true),
	}

	return f, nil
}

// Read function is used to implement the io.Reader interface. This can be used
// to read data as a stream. Read is much slower than ReadAt because only one
// read operation may run at a time. It uses ReadAt with stored offset.
func (f *File) Read(p []byte) (n int, err error) {
	if !f.open.Get() {
		return 0, goerr.Wrap(ErrClosed, 0)
	}

	f.rwmx.Lock()
	off := f.offs
	n, err = f.ReadAt(p, off)
	f.offs += int64(n)
	f.rwmx.Unlock()

	return n, goerr.Wrap(err, 0)
}

// ReadAt function is used to implement the io.ReaderAt interface. This will
// copy as many bytes from the memory map (starting from offset) to slice.
func (f *File) ReadAt(p []byte, off int64) (n int, err error) {
	if !f.open.Get() {
		return 0, goerr.Wrap(ErrClosed, 0)
	}

	f.iomx.RLock()

	fsz := atomic.LoadInt64(&f.size)
	end := off + int64(len(p))

	if end > fsz {
		copy(p, f.data.Data[off:fsz])
		n, err = int(fsz-off), io.EOF
	} else {
		copy(p, f.data.Data[off:end])
		n, err = int(end-off), nil
	}

	f.iomx.RUnlock()
	if err != nil {
		return n, goerr.Wrap(err, 0)
	}

	return n, nil
}

// Write function is used to implement the io.Writer interface. This can be
// used to write data as a stream. Write is much slower than WriteAt because
// only one write operation may run at a time. It uses WriteAt with stored
// offset. The file and memory map will grow automatically when necessary.
func (f *File) Write(p []byte) (n int, err error) {
	if !f.open.Get() {
		return 0, goerr.Wrap(ErrClosed, 0)
	}

	f.rwmx.Lock()
	off := f.offs
	n, err = f.WriteAt(p, off)
	f.offs += int64(n)
	f.rwmx.Unlock()

	return n, goerr.Wrap(err, 0)
}

// WriteAt function is used to implement the io.WriterAt interface. This will
// copy as many bytes from the slice to the memory map. The file and the memory
// map will grow automatically when necessary but it's recommended to allocate
// required space before using in order to write faster (ex. segfile package).
func (f *File) WriteAt(p []byte, off int64) (n int, err error) {
	if !f.open.Get() {
		return 0, goerr.Wrap(ErrClosed, 0)
	}

	end := off + int64(len(p))

	if fsz := atomic.LoadInt64(&f.size); fsz > end {
		f.iomx.RLock()
		n = copy(f.data.Data[off:end], p)
		f.iomx.RUnlock()
		return n, nil
	}

	f.iomx.Lock()
	defer f.iomx.Unlock()

	// write the data directly to the file
	// this will also increase the file size
	if n, err := f.file.WriteAt(p, off); err != nil {
		return n, goerr.Wrap(err, 0)
	} else if n != len(p) {
		return n, goerr.Wrap(fsutils.ErrWriteSz, 0)
	}

	if err := f.data.Close(); err != nil {
		return 0, goerr.Wrap(err, 0)
	}

	data, err := NewMMap(f.path, f.lock)
	if err != nil {
		return 0, goerr.Wrap(err, 0)
	}

	f.data = data
	f.size = int64(len(f.data.Data))

	return len(p), nil
}

// Size returns the size of the memory map. The file size can be different
// if the mapped file is written by some other program.
func (f *File) Size() (sz int64) {
	f.iomx.RLock()
	sz = f.size
	f.iomx.RUnlock()
	return sz
}

// Reset sets io.Reader/io.Writer offsets to the beginning of the file
func (f *File) Reset() {
	if !f.open.Get() {
		return
	}

	f.rwmx.Lock()
	f.offs = 0
	f.rwmx.Unlock()
}

// Sync synchronizes the memory map with the mapped file. This can ensure that
// all the data have been successfully and completely written to the disk.
func (f *File) Sync() (err error) {
	if !f.open.Get() {
		return goerr.Wrap(ErrClosed, 0)
	}

	f.iomx.Lock()
	defer f.iomx.Unlock()

	if err := f.data.Sync(); err != nil {
		return goerr.Wrap(err, 0)
	}

	return nil
}

// Close releases all resources
func (f *File) Close() (err error) {
	// R/W operations
	f.rwmx.Lock()
	f.iomx.Lock()
	defer f.rwmx.Unlock()
	defer f.iomx.Unlock()

	if !f.open.Get() {
		return goerr.Wrap(ErrClosed, 0)
	}

	f.open.Set(false)

	if err := f.data.Close(); err != nil {
		return goerr.Wrap(err, 0)
	}

	if err := f.file.Close(); err != nil {
		return goerr.Wrap(err, 0)
	}

	return nil
}
