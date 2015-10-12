package memmap

import (
	"errors"
	"os"
	"reflect"
	"syscall"
	"unsafe"
)

const (
	fmode = os.O_RDWR | os.O_CREATE
	fperm = 0644
	mflag = syscall.MAP_SHARED
	mprot = syscall.PROT_READ | syscall.PROT_WRITE
	msync = syscall.MS_SYNC
)

var (
	// ErrZeroSz is used when the user attempts to create a memory map
	// with zero file size. Provide a value > 0 for the size parameter.
	ErrZeroSz = errors.New("cannot create mmap with empty file")

	// ErrBadSz is used when the user attempts to create a memory map
	// with an existing file but its size if not equal to expected size.
	ErrBadSz = errors.New("cannot create mmap with empty file")
)

// Map is a struct which abstracts memory map system calls and provides a fast
// and easy to use api. The Map should be unmapped when not in use.
type Map struct {
	Data []byte
	hlen uintptr
	hadr uintptr
}

// New creates a new memory map struct on given path
// A new file will be created on given path if necessary.
// The file will be truncated to given size if it's empty.
func New(path string, size int64) (m *Map, err error) {
	if size == 0 {
		return nil, ErrZeroSz
	}

	file, err := os.OpenFile(path, fmode, fperm)
	if err != nil {
		return nil, err
	}

	// don't need this
	defer file.Close()

	m, err = MapFile(file, size)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// MapFile creates a new memory map struct from an os.File
// The file will be truncated to given size if it's empty.
func MapFile(file *os.File, size int64) (m *Map, err error) {
	if size == 0 {
		return nil, ErrZeroSz
	}

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if sz := info.Size(); sz != size {
		if sz != 0 {
			// file already exists with different size
			// this can be caused by corrupted files
			return nil, ErrBadSz
		}

		// If the file size if zero, it should be a new
		// segment file. Truncate it to required size.
		if err := file.Truncate(size); err != nil {
			return nil, err
		}

		sz = size
	}

	fd := file.Fd()
	data, err := syscall.Mmap(int(fd), 0, int(size), mprot, mflag)
	if err != nil {
		return nil, err
	}

	// get slice header to get memory address and length
	head := (*reflect.SliceHeader)(unsafe.Pointer(&data))

	m = &Map{
		Data: data,
		hadr: head.Data,
		hlen: uintptr(head.Len),
	}

	return m, nil
}

// Lock loads all memory pages in physical memory. This can take a long time for
// larger files but access to these memory locations will be faster.
func (m *Map) Lock() (err error) {
	if err := syscall.Mlock(m.Data); err != nil {
		return err
	}

	return nil
}

// Sync synchronizes the memory map with the mapped file. This can be used to
// ensure that all data is written to the disk successfully. Calling the Sync
// method is necessary to survive OS kernel level panics and crashes.
func (m *Map) Sync() (err error) {
	_, _, errno := syscall.Syscall(syscall.SYS_MSYNC, m.hadr, m.hlen, msync)
	if errno != 0 {
		err := syscall.Errno(errno)
		return err
	}

	return nil
}

// Close unmaps data and closes the file handler. Changes done to the memory
// map will be synced to the disk before closing to prevent data loss.
func (m *Map) Close() (err error) {
	if err := m.Sync(); err != nil {
		return err
	}

	if err := syscall.Munmap(m.Data); err != nil {
		return err
	}

	return nil
}
