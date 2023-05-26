// Package ir implements interfaces for the encoding and decoding of [CLP] IR
// (intermediate representation) streams through CLP's FFI (foreign function
// interface). More details on CLP IR streams are described in this [Uber
// blog].
// Log events compressed in IR format can be viewed in the [log viewer] or
// programmatically analyzed using APIs provided here. They can also be
// decompressed back into plain-text log files using CLP (in a future release).
//
// [CLP]: https://github.com/y-scope/clp
// [Uber blog]: https://www.uber.com/blog/reducing-logging-cost-by-two-orders-of-magnitude-using-clp/
// [log viewer]: https://github.com/y-scope/yscope-log-viewer
package ir

import (
	"unsafe"

	"github.com/y-scope/clp-ffi-go/ffi"
)

// TimestampInfo contains information relevant to all timestamps in the IR
// stream. This information comes from the metadata in the IR preamble.
type TimestampInfo struct {
	Pattern       string
	PatternSyntax string
	TimeZoneId    string
}

// Empty types used to constrain irStream to ensure the correct encoding size
// is used during encoding and decoding.
type (
	EightByteEncodedVariable struct{}
	FourByteEncodedVariable  struct{}
	EncodedVariable          interface {
		EightByteEncodedVariable | FourByteEncodedVariable
	}
)

// irStream is constrained by EncodedVariable to prevent mistaken usage of an
// incorrect sized stream.
type irStream[T EncodedVariable] struct {
	tsInfo TimestampInfo
	cPtr   unsafe.Pointer // currently unused in the decoder path
}

// Returns the TimestampInfo of an irStream.
func (self irStream[T]) TimestampInfo() TimestampInfo {
	return self.tsInfo
}

// Returns the TimestampInfo of an irStream.
type EightByteIrStream struct {
	irStream[EightByteEncodedVariable]
}

// FourByteIrStream contains both a CLP IR stream (irStream) and keeps track of
// the previous timestamp seen in the stream. Four byte encoding encodes log
// event timestamps as time deltas from the previous log event. Therefore, we
// must track the previous timestamp to be able to calculate the full timestamp
// of a log event.
type FourByteIrStream struct {
	irStream[FourByteEncodedVariable]
	prevTimestamp ffi.EpochTimeMs
}
