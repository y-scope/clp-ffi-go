// The ffi package defines types for log events used in logging functions and
// libraries, without CLP encoding or serialization.
package ffi

// LogEvent provides programmatic access to the various components of a log
// event.
type LogEvent struct {
	AutoKvPairs map[string]any
	UserKvPairs map[string]any
}

type MsgPackLogEvent struct {
	AutoKvPairs []byte
	UserKvPairs []byte
}

func NewLogEvent() *LogEvent {
	return &LogEvent {
		AutoKvPairs: make(map[string]any),
		UserKvPairs: make(map[string]any),
	}
}
