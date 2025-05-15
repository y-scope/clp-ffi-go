//go:build !external && arm64

package ir

/*
#cgo CPPFLAGS: -I${SRCDIR}/../pre-built/include/
#cgo linux LDFLAGS: ${SRCDIR}/../pre-built/lib/libclp-ffi-go_linux_arm64.a -lstdc++ -lm
#cgo darwin LDFLAGS: ${SRCDIR}/../pre-built/lib/libclp-ffi-go_darwin_arm64.a -lstdc++ -lm
*/
import "C"
