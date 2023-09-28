//go:build external

// When using `external` build manually set linkage with `CGO_LDFLAGS`.
package search

/*
#cgo CPPFLAGS: -I${SRCDIR}/../include/
#cgo external LDFLAGS:
*/
import "C"
