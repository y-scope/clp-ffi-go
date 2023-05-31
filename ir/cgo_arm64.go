//go:build !external && arm64

package ir

/*
#cgo CFLAGS: -I${SRCDIR}/../cpp/src/
#cgo linux LDFLAGS: -L${SRCDIR}/../lib/ -lclp_ffi_linux_arm64 -Wl,-rpath=${SRCDIR}/../lib/
#cgo darwin LDFLAGS: -L${SRCDIR}/../lib/ -lclp_ffi_darwin_arm64 -Wl,-rpath=${SRCDIR}/../lib/
*/
import "C"
