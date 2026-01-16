//go:generate ../scripts/download-libs.sh

package ir

/*
#include <clp_ffi_go/defs.h>
*/
import "C"

import (
	"unsafe"
)

// The follow functions are helpers to cleanup Cgo related code. The underlying
// Go type created from a 'C' type is not exported and recreated in each
// package. Therefore, these helpers must be redefined in any package wishing to
// use them, so that they reference the correct underlying Go type of the
// package (see: https://pkg.go.dev/cmd/cgo). This problem could be alivated by
// using Go generate to create/add these helpers to a package necessary.

func newCByteSpan(s []byte) C.ByteSpan {
	return C.ByteSpan{
		unsafe.Pointer(unsafe.SliceData(s)),
		C.size_t(len(s)),
	}
}
