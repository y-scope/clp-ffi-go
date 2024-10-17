package ir

/*
#include <ffi_go/defs.h>
#include <ffi_go/ir/deserializer.h>
*/
import "C"

import (
	"unsafe"

	"github.com/y-scope/clp-ffi-go/ffi"
)

// A Deserializer exports functions to deserialize log events from a CLP IR byte
// stream. Deserialization functions take an IR buffer as input, but how that
// buffer is materialized is left to the user. These functions return views
// (slices) of the log events extracted from the IR. Each Deserializer owns its
// own unique underlying memory for the views it produces/returns. This memory
// is reused for each view, so to persist the contents the memory must be copied
// into another object. Close must be called to free the underlying memory and
// failure to do so will result in a memory leak.
type Deserializer struct {
	cptr unsafe.Pointer
}

// DeserializePreamble attempts to read an IR stream preamble from irBuf,
// returning an Deserializer (of the correct stream encoding size), the position
// read to in irBuf (the end of the preamble), and an error. Note the metadata
// stored in the preamble is sparse and certain fields in TimestampInfo may be 0
// value. On error returns:
//   - nil Deserializer
//   - 0 position
//   - [IrError] error: CLP failed to successfully deserialize
//   - [encoding/json] error: unmarshalling the metadata failed
func DeserializePreamble(irBuf []byte) (*Deserializer, int, error) {
	if 0 >= len(irBuf) {
		return nil, 0, IncompleteIr
	}

	// TODO: Add version validation in this method or ir_deserializer_new_deserializer_with_preamble
	// after updating the clp version.

	var pos C.size_t
	var deserializerCptr unsafe.Pointer
	if err := IrError(C.ir_deserializer_create(
		newCByteSpan(irBuf),
		&pos,
		&deserializerCptr,
	)); Success != err {
		return nil, int(pos), err
	}

	return &Deserializer{deserializerCptr}, int(pos), nil
}

// Close will delete the underlying C++ allocated memory used by the
// deserializer. Failure to call Close will result in a memory leak.
func (deserializer *Deserializer) Close() error {
	if nil != deserializer.cptr {
		C.ir_deserializer_close(deserializer.cptr)
		deserializer.cptr = nil
	}
	return nil
}

// DeserializeLogEvent attempts to read the next log event from the IR stream in
// irBuf, returning the deserialized [ffi.LogEventView], the position read to in
// irBuf (the end of the log event in irBuf), and an error. On error returns:
//   - nil *ffi.LogEventView
//   - 0 position
//   - [IrError] error: CLP failed to successfully deserialize
//   - [EndOfIr] error: CLP found the IR stream EOF tag
func (deserializer *Deserializer) DeserializeLogEvent(
	irBuf []byte,
) (ffi.LogEvent, int, error) {
	return deserializeLogEvent(deserializer, irBuf)
}

func deserializeLogEvent(
	deserializer *Deserializer,
	irBuf []byte,
) (ffi.LogEvent, int, error) {
	if 0 >= len(irBuf) {
		return nil, 0, IncompleteIr
	}

	var pos C.size_t
	var msgpack_log_event C.ByteSpan
	var err error = IrError(C.ir_deserializer_deserialize_log_event(
			newCByteSpan(irBuf),
			deserializer.cptr,
			&pos,
			&msgpack_log_event,
		))
	if Success != err {
		return nil, 0, err
	}

	return nil, 0, nil
	// return &ffi.LogEventView{
	// 		LogMessageView: unsafe.String(
	// 			(*byte)((unsafe.Pointer)(event.m_log_message.m_data)),
	// 			event.m_log_message.m_size,
	// 		),
	// 		Timestamp: ffi.EpochTimeMs(event.m_timestamp),
	// 	},
	// 	int(pos),
	// 	nil
}
