package ir

// IRError mirrors cpp type IRErrorCode defined in:
// clp/components/core/src/ffi/ir_stream/decoding_methods.hpp
//go:generate stringer -type=IRError
type IRError int

const (
	Success IRError = iota
	DecodeError
	Eof
	CorruptedIR
	CorruptedMetadata
	IncompleteIR
	UnsupportedVersion
)

func (self IRError) Error() string {
	return self.String()
}
