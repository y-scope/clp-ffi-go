//go:build external

// When using `external` build manually set linkage with `CGO_LDFLAGS`.
package ir

/*
#cgo CPPFLAGS: -I${SRCDIR}/../pre-built/include/
#cgo external LDFLAGS:
*/
import "C"
