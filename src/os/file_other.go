//go:build baremetal || (tinygo.wasm && !wasip1 && !wasip2) || nintendoswitch

package os

import (
	_ "unsafe"
)

// Stdin, Stdout, and Stderr are open Files pointing to the standard input,
// standard output, and standard error file descriptors.
var (
	Stdin  = NewFile(0, "/dev/stdin")
	Stdout = NewFile(1, "/dev/stdout")
	Stderr = NewFile(2, "/dev/stderr")
)

const DevNull = "/dev/null"

// isOS indicates whether we're running on a real operating system with
// filesystem support.
const isOS = false

// stdioFileHandle represents one of stdin, stdout, or stderr depending on the
// number. It implements the FileHandle interface.
type stdioFileHandle uint8

// file is the real representation of *File.
// The extra level of indirection ensures that no clients of os
// can overwrite this data, which could cause the finalizer
// to close the wrong file descriptor.
type file struct {
	handle     FileHandle
	name       string
	appendMode bool
}

func (f *file) close() error {
	return f.handle.Close()
}

func NewFile(fd uintptr, name string) *File {
	return &File{&file{handle: stdioFileHandle(fd), name: name}}
}

// Chdir changes the current working directory to the named directory.
// If there is an error, it will be of type *PathError.
func Chdir(dir string) error {
	return ErrNotImplemented
}

// Rename renames (moves) oldpath to newpath.
// If newpath already exists and is not a directory, Rename replaces it.
// OS-specific restrictions may apply when oldpath and newpath are in different directories.
// If there is an error, it will be of type *LinkError.
func Rename(oldpath, newpath string) error {
	return ErrNotImplemented
}

// Read reads up to len(b) bytes from machine.Serial.
// It returns the number of bytes read and any error encountered.
func (f stdioFileHandle) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}

	size := buffered()
	for size == 0 {
		gosched()
		size = buffered()
	}

	if size > len(b) {
		size = len(b)
	}
	for i := 0; i < size; i++ {
		b[i] = getchar()
	}
	return size, nil
}

func (f stdioFileHandle) ReadAt(b []byte, off int64) (n int, err error) {
	return 0, ErrNotImplemented
}

func (f stdioFileHandle) WriteAt(b []byte, off int64) (n int, err error) {
	return 0, ErrNotImplemented
}

// Write writes len(b) bytes to the output. It returns the number of bytes
// written or an error if this file is not stdout or stderr.
func (f stdioFileHandle) Write(b []byte) (n int, err error) {
	switch f {
	case 1, 2: // stdout, stderr
		for _, c := range b {
			putchar(c)
		}
		return len(b), nil
	default:
		return 0, ErrUnsupported
	}
}

// Close is unsupported on this system.
func (f stdioFileHandle) Close() error {
	return ErrUnsupported
}

// Seek wraps syscall.Seek.
func (f stdioFileHandle) Seek(offset int64, whence int) (int64, error) {
	return -1, ErrUnsupported
}

func (f stdioFileHandle) Sync() error {
	return ErrUnsupported
}

func (f stdioFileHandle) Fd() uintptr {
	return uintptr(f)
}

//go:linkname putchar runtime.putchar
func putchar(c byte)

//go:linkname getchar runtime.getchar
func getchar() byte

//go:linkname buffered runtime.buffered
func buffered() int

//go:linkname gosched runtime.Gosched
func gosched() int

func Pipe() (r *File, w *File, err error) {
	return nil, nil, ErrNotImplemented
}

func Symlink(oldname, newname string) error {
	return ErrNotImplemented
}

func Readlink(name string) (string, error) {
	return "", ErrNotImplemented
}

func tempDir() string {
	return "/tmp"
}

// Truncate is unsupported on this system.
func Truncate(filename string, size int64) (err error) {
	return ErrUnsupported
}

// Truncate is unsupported on this system.
func (f *File) Truncate(size int64) (err error) {
	if f.handle == nil {
		return ErrClosed
	}

	return Truncate(f.name, size)
}

func (f *File) chmod(mode FileMode) error {
	return ErrUnsupported
}

func (f *File) chdir() error {
	return ErrNotImplemented
}
