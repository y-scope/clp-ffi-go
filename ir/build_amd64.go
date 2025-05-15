//go:build !external && amd64

package ir

/*
#cgo CPPFLAGS: -I${SRCDIR}/../pre-built/include/
#cgo linux LDFLAGS: ${SRCDIR}/../pre-built/lib/libclp-ffi-go_linux_amd64.a -lstdc++ -lm
#cgo darwin LDFLAGS: ${SRCDIR}/../pre-built/lib/libclp-ffi-go_darwin_amd64.a -lstdc++ -lm
*/
import "C"
