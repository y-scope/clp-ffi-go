package ir

/*
#include "../cpp/src/log_event.h"
#include "../cpp/src/ir/decoding.h"
*/
import "C"

import (
	"encoding/json"
	"strconv"
	"unsafe"

	"github.com/y-scope/clp-ffi-go" // ffi::types + necessary for linkage
)

// TODO once we reach go >= 1.20
// change &buf[:1][0] to unsafe.SliceData(buf)
// https://pkg.go.dev/unsafe#SliceData

// IrDecoder exports functions to decode log events in an IR stream and also
// inspect the timnestamp information of the stream. An IrDecoder manages the
// internal state of the IR stream, such that the next log event in the stream
// can be decoded. The maintence of the buffer containing the IR stream is left
// to the caller.
type IrDecoder interface {
	DecodeNextLogEvent(buf []byte) (ffi.LogEvent, int, error)
	TimestampInfo() TimestampInfo
}

// DecodePreamble attempts to read an IR stream preamble from buf, returning an
// IrDecoder (of the correct stream encoding size), the offset read to in buf
// (the end of the preamble), and an error. Note the metadata stored in the
// preamble is sparse and certain fields in TimestampInfo may be 0 value.
// Return values:
//   - nil == error: successful decode
//   - nil != error: IrDecode will be nil, offset may be non-zero for debugging purposes
//     - type [IRError]: CLP failed to successfully decode
//     - type from [encoding/json]: unmarshalling the metadata failed
func DecodePreamble(buf []byte) (IrDecoder, int, error) {
	var offset C.size_t
	var ir_encoding C.int8_t
	var metadata_type C.int8_t
	var metadata_pos C.size_t
	var metadata_size C.uint16_t

	if err := IRError(C.decode_preamble(
		unsafe.Pointer(&buf[:1][0]),
		C.size_t(len(buf)),
		&offset,
		&ir_encoding,
		&metadata_type,
		&metadata_pos,
		&metadata_size)); Success != err {
		return nil, int(offset), err
	}

	if 1 != metadata_type {
		return nil, int(offset), UnsupportedVersion
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(buf[metadata_pos:metadata_pos+C.size_t(metadata_size)], &metadata); nil != err {
		return nil, int(offset), err
	}

	var tsInfo TimestampInfo
	if tsPat, ok := metadata["TIMESTAMP_PATTERN"].(string); ok {
		tsInfo.Pattern = tsPat
	}
	if tsSyn, ok := metadata["TIMESTAMP_PATTERN_SYNTAX"].(string); ok {
		tsInfo.PatternSyntax = tsSyn
	}
	if tzid, ok := metadata["TZ_ID"].(string); ok {
		tsInfo.TimeZoneId = tzid
	}

	var decoder IrDecoder
	if 1 == ir_encoding {
		var refTs ffi.EpochTimeMs = 0
		if tsStr, ok := metadata["REFERENCE_TIMESTAMP"].(string); ok {
			if tsInt, err := strconv.ParseInt(tsStr, 10, 64); nil == err {
				refTs = ffi.EpochTimeMs(tsInt)
			}
		}
		decoder = &FourByteIrStream{
			irStream:      irStream[FourByteEncodedVariable]{tsInfo: tsInfo, cPtr: nil},
			prevTimestamp: refTs,
		}
	} else {
		decoder = &EightByteIrStream{irStream[EightByteEncodedVariable]{tsInfo, nil}}
	}

	return decoder, int(offset), nil
}

// DecodeNextLogEvent attempts to read the next LogEvent from the IR stream in
// buf, returning the LogEvent, the offset read to in buf (the end of the
// LogEvent in buf), and an error.
// Return values:
//   - nil == error: successful decode
//   - nil != error: ffi.LogEvent will be nil, offset may be non-zero for debugging purposes
//     - io.EOF: CLP found the IR stream EOF tag
//     - type IRError: CLP failed to successfully decode
func (self *EightByteIrStream) DecodeNextLogEvent(buf []byte) (ffi.LogEvent, int, error) {
	return decodeNextLogEvent(self, buf)
}

// DecodeNextLogEvent attempts to read the next LogEvent from the IR stream in
// buf, returning the LogEvent, the offset read to in buf (the end of the
// LogEvent in buf), and an error.
// Return values:
//   - nil == error: successful decode
//   - nil != error: ffi.LogEvent will be nil, offset may be non-zero for debugging purposes
//     - [IRError.Eof] -> CLP found the IR stream EOF tag
//     - type IRError -> CLP failed to successfully decode
func (self *FourByteIrStream) DecodeNextLogEvent(buf []byte) (ffi.LogEvent, int, error) {
	return decodeNextLogEvent(self, buf)
}

// decodeNextLogEvent performs the actual work for DecodeNextLogEvent in a
// generic way.
func decodeNextLogEvent[T EightByteIrStream | FourByteIrStream](
	irstream *T,
	buf []byte,
) (ffi.LogEvent, int, error) {
	if 0 >= len(buf) {
		return ffi.LogEvent{}, 0, IncompleteIR
	}
	var offset C.size_t
	var msgObj unsafe.Pointer
	var msg *C.char
	var msgSize C.size_t
	var timestampOrDelta C.int64_t

	var err error
	switch any(irstream).(type) {
	case *EightByteIrStream:
		err = IRError(C.eight_byte_decode_next_log_event(
			unsafe.Pointer(&buf[:1][0]),
			C.size_t(len(buf)),
			&offset,
			&msgObj,
			&msg,
			&msgSize,
			&timestampOrDelta))
	case *FourByteIrStream:
		err = IRError(C.four_byte_decode_next_log_event(
			unsafe.Pointer(&buf[:1][0]),
			C.size_t(len(buf)),
			&offset,
			&msgObj,
			&msg,
			&msgSize,
			&timestampOrDelta))
	default:
		return ffi.LogEvent{}, 0, UnsupportedVersion
	}
	if Success != err {
		return ffi.LogEvent{}, int(offset), err
	}

	var ts ffi.EpochTimeMs
	switch irs := any(irstream).(type) {
	case *EightByteIrStream:
		ts = ffi.EpochTimeMs(timestampOrDelta)
	case *FourByteIrStream:
		ts = irs.prevTimestamp + ffi.EpochTimeMs(timestampOrDelta)
		irs.prevTimestamp = ts
	default:
		return ffi.LogEvent{}, 0, UnsupportedVersion
	}

	event := ffi.LogEvent{
		LogMessage: ffi.NewLogMessage(unsafe.Pointer(msg), uint64(msgSize), msgObj),
		Timestamp:  ts,
	}
	return event, int(offset), nil
}
