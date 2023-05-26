//go:build external

// When using `external` build manually set linkage with `CGO_LDFLAGS`.
package ffi

/*
#cgo CFLAGS: -I${SRCDIR}/../cpp/src
#cgo external LDFLAGS:
*/
import "C"
