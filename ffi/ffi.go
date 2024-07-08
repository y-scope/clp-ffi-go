// The ffi package defines types for log events used in logging functions and
// libraries, without CLP encoding or serialization.
package ffi

// Mirrors cpp type epoch_time_ms_t
type EpochTimeMs int64

// A ffi.LogMessage represents the text (message) component of a log event.
// A LogMessageView is a LogMessage that is backed by C++ allocated memory
// rather than the Go heap. A LogMessageView, x, is valid when returned and will
// remain valid until a new LogMessageView is returned by the same object (e.g.
// an ir.Deserializer) that returns x.
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

// LogEventView memory is allocated and owned by the C++ object (e.g., reader,
// deserializer) that returns it. Reusing a LogEventView after the same object
// has issued a new view leads to undefined behavior, as different objects
// manage their own memory independently.
type LogEventView struct {
	LogMessageView
	Timestamp EpochTimeMs
	UtcOffset EpochTimeMs
}
