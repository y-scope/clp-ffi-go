//go:build external

// When using `external` build manually set linkage with `CGO_LDFLAGS`.
package ir

/*
#cgo CPPFLAGS: -I${SRCDIR}/../include/
#cgo external LDFLAGS:
*/
import "C"
