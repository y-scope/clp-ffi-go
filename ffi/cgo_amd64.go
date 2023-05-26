//go:build !external && amd64

package ffi

/*
#cgo CFLAGS: -I${SRCDIR}/../cpp/src/
#cgo linux LDFLAGS: -L${SRCDIR}/../lib/ -lclp_ffi_linux_amd64 -lstdc++
#cgo darwin LDFLAGS: -L${SRCDIR}/../lib/ -lclp_ffi_darwin_amd64 -lstdc++
*/
import "C"
