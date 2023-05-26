//go:build !external && arm64

package message

/*
#cgo CFLAGS: -I${SRCDIR}/../cpp/src
#cgo linux LDFLAGS: -L${SRCDIR}/../lib -lclp_ffi_linux_arm64 -lstdc++
#cgo darwin LDFLAGS: -L${SRCDIR}/../lib -lclp_ffi_darwin_arm64 -lstdc++
*/
import "C"
