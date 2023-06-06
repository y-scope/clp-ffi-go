// The ffi package contains the general types representing log events  created
// and used by logging functions or libraries. In other words log events with no
// sort of CLP encoding or serializing.
package ffi

// Mirrors cpp type epoch_time_ms_t
type EpochTimeMs int64

// A ffi.LogMessage represents the text (message) component of a log event.
// A LogMessageView is a LogMessage that is backed by C++ allocated memory
// rather than the Go heap. A LogMessageView, x, is valid when returned and will
// remain valid until a new LogMessageView is returned by the same object (e.g.
// an ir.Deserializer) that retuend x.
type (
	LogMessageView = string
	LogMessage     = string
)

// LogEvent provides programmatic access to the various components of a log
// event.
type LogEvent struct {
	LogMessage
	Timestamp EpochTimeMs
}

// The underlying memory of LogEventView is C-allocated and owned by the object
// (e.g. reader, desializer, etc) that returned it. Using an existing
// LogEventView after a new view has been returned by the same producing object
// is undefined (different producing objects own their own memory for views).
type LogEventView struct {
	LogMessageView
	Timestamp EpochTimeMs
}
