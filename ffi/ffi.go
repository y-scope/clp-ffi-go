// The ffi package defines types for log events used in logging functions and
// libraries, without CLP encoding or serialization.
package ffi

// LogEvent provides programmatic access to the various components of a log
// event.
type LogEvent = map[string]interface{}
