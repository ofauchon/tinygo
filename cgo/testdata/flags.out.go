// CGo errors:
//     testdata/flags.go:5:7: invalid #cgo line: NOFLAGS
//     testdata/flags.go:8:13: invalid flag: -fdoes-not-exist
//     testdata/flags.go:29:14: invalid flag: -does-not-exists

package main

import "syscall"
import "unsafe"

var _ unsafe.Pointer

//go:linkname _Cgo_CString runtime.cgo_CString
func _Cgo_CString(string) *_Cgo_char

//go:linkname _Cgo_GoString runtime.cgo_GoString
func _Cgo_GoString(*_Cgo_char) string

//go:linkname _Cgo___GoStringN runtime.cgo_GoStringN
func _Cgo___GoStringN(*_Cgo_char, uintptr) string

func _Cgo_GoStringN(cstr *_Cgo_char, length _Cgo_int) string {
	return _Cgo___GoStringN(cstr, uintptr(length))
}

//go:linkname _Cgo___GoBytes runtime.cgo_GoBytes
func _Cgo___GoBytes(unsafe.Pointer, uintptr) []byte

func _Cgo_GoBytes(ptr unsafe.Pointer, length _Cgo_int) []byte {
	return _Cgo___GoBytes(ptr, uintptr(length))
}

//go:linkname _Cgo___CBytes runtime.cgo_CBytes
func _Cgo___CBytes([]byte) unsafe.Pointer

func _Cgo_CBytes(b []byte) unsafe.Pointer {
	return _Cgo___CBytes(b)
}

//go:linkname _Cgo___get_errno_num runtime.cgo_errno
func _Cgo___get_errno_num() uintptr

func _Cgo___get_errno() error {
	return syscall.Errno(_Cgo___get_errno_num())
}

type (
	_Cgo_char      uint8
	_Cgo_schar     int8
	_Cgo_uchar     uint8
	_Cgo_short     int16
	_Cgo_ushort    uint16
	_Cgo_int       int32
	_Cgo_uint      uint32
	_Cgo_long      int32
	_Cgo_ulong     uint32
	_Cgo_longlong  int64
	_Cgo_ulonglong uint64
)

const _Cgo_BAR = 3
const _Cgo_FOO_H = 1
