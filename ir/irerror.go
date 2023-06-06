package ir

// IrError mirrors cpp type IRErrorCode defined in:
// clp/components/core/src/ffi/ir_stream/decoding_methods.hpp
//
//go:generate stringer -type=IrError
type IrError int

const (
	Success IrError = iota
	DecodeError
	EndOfIr
	CorruptedIr
	IncompleteIr
	QueryNotFound      // must be IncompleteIr + 1
	EncodeError        // not from clp
	UnsupportedVersion // not from clp
)

func (self IrError) Error() string {
	return self.String()
}
