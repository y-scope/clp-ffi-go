//go:build !external && amd64

package message

/*
#cgo CFLAGS: -I${SRCDIR}/../cpp/src/
#cgo linux LDFLAGS: -L${SRCDIR}/../lib/ -lclp_ffi_linux_amd64 -Wl,-rpath=${SRCDIR}/../lib/
#cgo darwin LDFLAGS: -L${SRCDIR}/../lib/ -lclp_ffi_darwin_amd64 -Wl,-rpath=${SRCDIR}/../lib/
*/
import "C"
