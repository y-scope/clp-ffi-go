// The ir package implements interfaces for the encoding, decoding,
// serialization, and deserialization of [CLP] IR (intermediate representation)
// streams through CLP's FFI (foreign function interface). More details on CLP
// IR streams are described in this [Uber blog].
// Log events in IR format can be viewed in the [log viewer] or programmatically
// analyzed using APIs provided in this package.
//
// [CLP]: https://github.com/y-scope/clp
// [Uber blog]: https://www.uber.com/blog/reducing-logging-cost-by-two-orders-of-magnitude-using-clp/
// [log viewer]: https://github.com/y-scope/yscope-log-viewer
package ir

/*
#include <ffi_go/defs.h>
*/
import "C"

// Must match c++ equivalent types
type (
	EightByteEncoding = int64
	FourByteEncoding  = int32
)

// TimestampInfo contains general information applying to all timestamps in
// contiguous IR. This information comes from the metadata in the IR preamble.
type TimestampInfo struct {
	Pattern       string
	PatternSyntax string
	TimeZoneId    string
}

// ir.BufView is a slice of CLP IR backed by C++ allocated memory rather than
// the Go heap. A BufView, x, is valid when returned and will remain valid until
// a new BufView is returned by the same object (e.g. an [ir.Serializer]) that
// retuend x.
type BufView = []byte

// A ir.LogMessage contains all the different components of a log message
// ([ffi.LogMessage]) encoded/separated into fields.
type LogMessage[T EightByteEncoding | FourByteEncoding] struct {
	Logtype           string
	Vars              []T
	DictVars          string
	DictVarEndOffsets []int32
}

// A ir.LogMessageView is a [ir.LogMessage] that is backed by C++ allocated
// memory rather than the Go heap. A LogMessageView, x, is valid when returned
// and will remain valid until a new LogMessageView is returned by the same
// object (e.g.  an [ir.Encoder]) that retuend x.
type LogMessageView[T EightByteEncoding | FourByteEncoding] struct {
	LogMessage[T]
}
