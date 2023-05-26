package ffi

/*
#include <log_event.h>
*/
import "C"

import (
	"runtime"
	"unsafe"
)

// Mirrors cpp type epoch_time_ms_t defined in:
// src/ir/encoding.h
// src/ir/decoding.h
type EpochTimeMs int64

type cppReference struct {
	cptr unsafe.Pointer
}

type LogMessage struct {
	Msg  []byte
	cref *cppReference
}

// Creates a new LogMessage backed by C-allocated memory and sets
// [finalizeLogMessage] as a finalizer.
func NewLogMessage (msg unsafe.Pointer, msgSize uint64, obj unsafe.Pointer) LogMessage {
	ref := &cppReference{obj}
	log := LogMessage{unsafe.Slice((*byte)(msg), msgSize), ref}
	runtime.SetFinalizer(ref, finalizeLogMessage)
	return log
}

// DeleteLogMessage calls down to C where any additional clean up occurs before
// calling delete on the stored class pointer. After calling this function log
// is in an empty/nil state and the finalizer is unset. This function is only
// useful if the memory overhead of relying on the finalizer to call delete is
// a concern.
func DeleteLogMessage(log *LogMessage) {
	if nil != log.cref {
		log.Msg = nil
		C.delete_log_event(log.cref.cptr)
		runtime.SetFinalizer(log.cref, nil)
		log.cref = nil
	}
}

// All LogMessages created with NewLogMessage will use this function as a
// finalizer to mimic GC. If memory overhead is a concern call
// [DeleteLogMessage] to immediately call delete (it will also clean up
// LogMessage and guards against double free).
//
// The rules for finalizers running are not perfectly equivalent to
// Go-allocated memory being GC'd, but in the case of LogMessages the
// C-allocated memory should eventually be deleted in similar fashion to a
// Go-allocated equivalent object. See
// https://pkg.go.dev/runtime#SetFinalizer.
func finalizeLogMessage(obj *cppReference) {
	if nil != obj {
		C.delete_log_event(obj.cptr)
	}
}

type LogEvent struct {
	LogMessage
	Timestamp EpochTimeMs
}
