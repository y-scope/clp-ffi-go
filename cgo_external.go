//go:build external

// When using `external` build manually set linkage with `CGO_LDFLAGS`. This
// file exists so that child packages can inherit this linking without
// duplicating all the files/logic.
package ffi

/*
#cgo external LDFLAGS:
*/
import "C"
