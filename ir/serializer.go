package ir

/*
#include <ffi_go/defs.h>
#include <ffi_go/ir/serializer.h>
*/
import "C"

import (
	"unsafe"

	"github.com/y-scope/clp-ffi-go/ffi"
)

// A Serializer exports functions to serialize log events into a CLP IR byte
// stream. Serialization functions only return views (slices) of IR bytes,
// leaving their use to the user. Each Serializer owns its own unique underlying
// memory for the views it produces/returns. This memory is reused for each
// view, so to persist the contents the memory must be copied into another
// object. Close must be called to free the underlying memory and failure to do
// so will result in a memory leak.
type Serializer interface {
	SerializeLogEvent(event ffi.LogEvent) (BufView, error)
	SerializeUtcOffsetChange(utcOffset ffi.EpochTimeMs) BufView
	TimestampInfo() TimestampInfo
	Close() error
}

// EightByteSerializer creates and returns a new Serializer that writes eight
// byte encoded CLP IR and serializes a IR preamble into a BufView using it. On
// error returns:
//   - nil Serializer
//   - nil BufView
//   - [IrError] error: CLP failed to successfully serialize
func EightByteSerializer(
	tsPattern string,
	tsPatternSyntax string,
) (Serializer, BufView, error) {
	var irView C.ByteSpan
	irs := eightByteSerializer{
		commonSerializer{TimestampInfo{tsPattern, tsPatternSyntax}, nil},
	}
	if err := IrError(C.ir_serializer_new_eight_byte_serializer_with_preamble(
		newCStringView(tsPattern),
		newCStringView(tsPatternSyntax),
		&irs.cptr,
		&irView,
	)); Success != err {
		return nil, nil, err
	}
	return &irs, unsafe.Slice((*byte)(irView.m_data), irView.m_size), nil
}

// FourByteSerializer creates and returns a new Serializer that writes four byte
// encoded CLP IR and serializes a IR preamble into a BufView using it. On error
// returns:
//   - nil Serializer
//   - nil BufView
//   - [IrError] error: CLP failed to successfully serialize
func FourByteSerializer(
	tsPattern string,
	tsPatternSyntax string,
	referenceTs ffi.EpochTimeMs,
) (Serializer, BufView, error) {
	var irView C.ByteSpan
	irs := fourByteSerializer{
		commonSerializer{TimestampInfo{tsPattern, tsPatternSyntax}, nil},
		referenceTs,
	}
	if err := IrError(C.ir_serializer_new_four_byte_serializer_with_preamble(
		newCStringView(tsPattern),
		newCStringView(tsPatternSyntax),
		C.int64_t(referenceTs),
		&irs.cptr,
		&irView,
	)); Success != err {
		return nil, nil, err
	}
	return &irs, unsafe.Slice((*byte)(irView.m_data), irView.m_size), nil
}

// commonSerializer contains fields common to all types of CLP IR encoding.
// TimestampInfo stores information common to all timestamps found in the IR.
// cptr holds a reference to the underlying C++ objected used as backing storage
// for the Views returned by the serializer. Close must be called to free this
// underlying memory and failure to do so will result in a memory leak.
type commonSerializer struct {
	tsInfo TimestampInfo
	cptr   unsafe.Pointer
}

// SerializeUtcOffsetChange serializes the UTC offset change, utcOffset, into an IR stream. It
// returns a view of the encoded IR bytes.
func (serializer *commonSerializer) SerializeUtcOffsetChange(utcOffset ffi.EpochTimeMs) BufView {
	var irView C.ByteSpan
	C.ir_serializer_serialize_utc_offset_change(C.int64_t(utcOffset), serializer.cptr, &irView)
	return unsafe.Slice((*byte)(irView.m_data), irView.m_size)
}

// Close attempts to close the serializer by releasing the underlying C++ allocated memory.
// Failure to call Close will result in a memory leak.
func (serializer *commonSerializer) Close() error {
	if nil != serializer.cptr {
		C.ir_serializer_close(serializer.cptr)
		serializer.cptr = nil
	}
	return nil
}

// Returns the TimestampInfo of the Serializer.
func (serializer commonSerializer) TimestampInfo() TimestampInfo {
	return serializer.tsInfo
}

type eightByteSerializer struct {
	commonSerializer
}

// SerializeLogEvent attempts to serialize the log event, event, into an eight
// byte encoded CLP IR byte stream. On error returns:
//   - a nil BufView
//   - [IrError] based on the failure of the Cgo call
func (serializer *eightByteSerializer) SerializeLogEvent(
	event ffi.LogEvent,
) (BufView, error) {
	return serializeLogEvent(serializer, event)
}

// fourByteSerializer contains both a common CLP IR serializer and stores the
// previously seen log event's timestamp. The previous timestamp is necessary to
// calculate the current timestamp as four byte encoding only encodes the
// timestamp delta between the current log event and the previous.
type fourByteSerializer struct {
	commonSerializer
	prevTimestamp ffi.EpochTimeMs
}

// SerializeLogEvent attempts to serialize the log event, event, into a four
// byte encoded CLP IR byte stream. On error returns:
//   - nil BufView
//   - [IrError] based on the failure of the Cgo call
func (serializer *fourByteSerializer) SerializeLogEvent(
	event ffi.LogEvent,
) (BufView, error) {
	return serializeLogEvent(serializer, event)
}

func serializeLogEvent(
	serializer Serializer,
	event ffi.LogEvent,
) (BufView, error) {
	var irView C.ByteSpan
	var err error
	switch irs := serializer.(type) {
	case *eightByteSerializer:
		err = IrError(C.ir_serializer_serialize_eight_byte_log_event(
			newCStringView(event.LogMessage),
			C.int64_t(event.Timestamp),
			irs.cptr,
			&irView,
		))
	case *fourByteSerializer:
		err = IrError(C.ir_serializer_serialize_four_byte_log_event(
			newCStringView(event.LogMessage),
			C.int64_t(event.Timestamp-irs.prevTimestamp),
			irs.cptr,
			&irView,
		))
		if Success == err {
			irs.prevTimestamp = event.Timestamp
		}
	}
	if Success != err {
		return nil, err
	}
	return unsafe.Slice((*byte)(irView.m_data), irView.m_size), nil
}
