//go:build !external && arm64

package ir

/*
#cgo CPPFLAGS: -I${SRCDIR}/../include/
#cgo linux LDFLAGS: ${SRCDIR}/../lib/libclp_ffi_linux_arm64.a -lstdc++
#cgo darwin LDFLAGS: ${SRCDIR}/../lib/libclp_ffi_darwin_arm64.a -lstdc++
*/
import "C"
