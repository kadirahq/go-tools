package fsutils

import (
	"errors"
	"os"
)

var (
	// ErrWrite is returned when bytes written is not equal to data size
	ErrWrite = errors.New("bytes written != data size")

	// ErrRead is returned when bytes read is not equal to data size
	ErrRead = errors.New("bytes read != data size")
)

// EnsureDir makes sure that a directory exists at path
// It will attempt to create a directory if not exists
func EnsureDir(dpath string) (err error) {
	// try to create target directory if missing
	return os.MkdirAll(dpath, 0755)
}

// EnsureFile amkes sure that a file exists at path with given Size
// Empty bytes will be appended if the file size is smaller
func EnsureFile(fpath string, sz int64) (err error) {
	file, err := os.OpenFile(fpath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	finfo, err := file.Stat()
	if err != nil {
		return err
	}

	initSize := finfo.Size()
	toAppend := sz - initSize

	// chunkData is a 1MB size slice of zeroes
	chunkSize := int64(1024 * 1024)
	chunkData := make([]byte, chunkSize)

	// number of chunks to write
	chunks := toAppend / chunkSize

	var i int64
	for i = 0; i < chunks; i++ {
		offset := initSize + chunkSize*i
		n, err := file.WriteAt(chunkData, offset)
		if err != nil {
			return err
		} else if int64(n) != chunkSize {
			return ErrWrite
		}
	}

	// write all remaining bytes
	toWrite := toAppend % chunkSize
	zeroes := chunkData[:toWrite]
	offset := chunkSize * chunks
	n, err := file.WriteAt(zeroes, offset)
	if err != nil {
		return err
	} else if int64(n) != toWrite {
		return ErrWrite
	}

	return nil
}
