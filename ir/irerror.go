package ir

// Mirrors the c++ type clp_ffi_go::ir::ErrorCode (cpp/src/clp_ffi_go/ir/ErrorCode.hpp) and is used
// when receiving an error code from cgo. The values of these types must match.
//
//go:generate stringer -type=IrError
type IrError int

const (
	Success IrError = iota
	InvalidArg
	IrBufferError
	IrCorrupt
	IrDecodeError
	IrEndOfStream
	IrIncomplete
	IrProtoBackwardCompatible
	IrProtoSupported
	IrProtoUnsupported
	IrSerializeError
	NotSupported
)

func (err IrError) Error() string {
	return err.String()
}
