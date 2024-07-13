package ir

/*
#include <ffi_go/defs.h>
#include <ffi_go/ir/deserializer.h>
#include <ffi_go/search/wildcard_query.h>
*/
import "C"

import (
	"encoding/json"
	"strconv"
	"unsafe"

	"github.com/y-scope/clp-ffi-go/ffi"
	"github.com/y-scope/clp-ffi-go/search"
)

const (
	metadataReferenceTimestampKey     = "REFERENCE_TIMESTAMP"
	metadataTimestampPatternKey       = "TIMESTAMP_PATTERN"
	metadataTimestampPatternSyntaxKey = "TIMESTAMP_PATTERN_SYNTAX"
)

// A Deserializer exports functions to deserialize log events from a CLP IR byte
// stream. Deserialization functions take an IR buffer as input, but how that
// buffer is materialized is left to the user. These functions return views
// (slices) of the log events extracted from the IR. Each Deserializer owns its
// own unique underlying memory for the views it produces/returns. This memory
// is reused for each view, so to persist the contents the memory must be copied
// into another object. Close must be called to free the underlying memory and
// failure to do so will result in a memory leak.
type Deserializer interface {
	DeserializeLogEvent(irBuf []byte) (*ffi.LogEventView, int, error)
	DeserializeWildcardMatchWithTimeInterval(
		irBuf []byte,
		mergedQuery search.MergedWildcardQuery,
		timeInterval search.TimestampInterval,
	) (*ffi.LogEventView, int, int, error)
	TimestampInfo() TimestampInfo
	Close() error
}

// DeserializePreamble attempts to read an IR stream preamble from irBuf,
// returning a Deserializer (of the correct stream encoding size), the position
// read to in irBuf (the end of the preamble), and an error. Note the metadata
// stored in the preamble is sparse and certain fields in TimestampInfo may be 0
// value. On error returns:
//   - nil Deserializer
//   - 0 position
//   - [IrError] error: CLP failed to successfully deserialize
//   - [encoding/json] error: unmarshalling the metadata failed
func DeserializePreamble(irBuf []byte) (Deserializer, int, error) {
	if 0 >= len(irBuf) {
		return nil, 0, IncompleteIr
	}

	// TODO: Add version validation in this method or ir_deserializer_new_deserializer_with_preamble
	// after updating the clp version.

	var pos C.size_t
	var irEncoding C.int8_t
	var metadataType C.int8_t
	var metadataPos C.size_t
	var metadataSize C.uint16_t
	var deserializerCptr unsafe.Pointer
	var timestampCptr unsafe.Pointer
	if err := IrError(C.ir_deserializer_new_deserializer_with_preamble(
		newCByteSpan(irBuf),
		&pos,
		&irEncoding,
		&metadataType,
		&metadataPos,
		&metadataSize,
		&deserializerCptr,
		&timestampCptr,
	)); Success != err {
		return nil, int(pos), err
	}

	if metadataType != 1 {
		return nil, 0, UnsupportedVersion
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(
		irBuf[metadataPos:metadataPos+C.size_t(metadataSize)],
		&metadata,
	); nil != err {
		return nil, 0, err
	}

	var tsInfo TimestampInfo
	if tsPat, ok := metadata[metadataTimestampPatternKey].(string); ok {
		tsInfo.Pattern = tsPat
	}
	if tsSyn, ok := metadata[metadataTimestampPatternSyntaxKey].(string); ok {
		tsInfo.PatternSyntax = tsSyn
	}

	var deserializer Deserializer
	if irEncoding == 1 {
		var refTs ffi.EpochTimeMs = 0
		if tsStr, ok := metadata[metadataReferenceTimestampKey].(string); ok {
			if tsInt, err := strconv.ParseInt(tsStr, 10, 64); nil == err {
				refTs = ffi.EpochTimeMs(tsInt)
				*(*ffi.EpochTimeMs)(timestampCptr) = refTs
			}
		}
		deserializer = &fourByteDeserializer{commonDeserializer{tsInfo, deserializerCptr}, refTs}
	} else {
		deserializer = &eightByteDeserializer{commonDeserializer{tsInfo, deserializerCptr}}
	}

	return deserializer, int(pos), nil
}

// commonDeserializer contains fields common to all types of CLP IR encoding.
// TimestampInfo stores information common to all timestamps found in the IR.
// cptr holds a reference to the underlying C++ objected used as backing storage
// for the Views returned by the deserializer. Close must be called to free this
// underlying memory and failure to do so will result in a memory leak.
type commonDeserializer struct {
	tsInfo TimestampInfo
	cptr   unsafe.Pointer
}

// Close will delete the underlying C++ allocated memory used by the
// deserializer. Failure to call Close will result in a memory leak.
func (deserializer *commonDeserializer) Close() error {
	if nil != deserializer.cptr {
		C.ir_deserializer_close(deserializer.cptr)
		deserializer.cptr = nil
	}
	return nil
}

// Returns the TimestampInfo used by the Deserializer.
func (deserializer commonDeserializer) TimestampInfo() TimestampInfo {
	return deserializer.tsInfo
}

type eightByteDeserializer struct {
	commonDeserializer
}

// DeserializeLogEvent attempts to read the next log event from the IR stream in
// irBuf, returning the deserialized [ffi.LogEventView], the position read to in
// irBuf (the end of the log event in irBuf), and an error. On error returns:
//   - nil *ffi.LogEventView
//   - 0 position
//   - [IrError] error: CLP failed to successfully deserialize
//   - [EndOfIr] error: CLP found the IR stream EOF tag
func (deserializer *eightByteDeserializer) DeserializeLogEvent(
	irBuf []byte,
) (*ffi.LogEventView, int, error) {
	return deserializeLogEvent(deserializer, irBuf)
}

// DeserializeWildcardMatchWithTimeInterval attempts to read the next log event
// from the IR stream in irBuf that matches mergedQuery within timeInterval. It
// returns the deserialized [ffi.LogEventView], the position read to in irBuf
// (the end of the log event in irBuf), the index of the matched query in
// mergedQuery, and an error. On error returns:
//   - nil *ffi.LogEventView
//   - 0 position
//   - -1 index
//   - [IrError] error: CLP failed to successfully deserialize
//   - [EndOfIr] error: CLP found the IR stream EOF tag
func (deserializer *eightByteDeserializer) DeserializeWildcardMatchWithTimeInterval(
	irBuf []byte,
	mergedQuery search.MergedWildcardQuery,
	timeInterval search.TimestampInterval,
) (*ffi.LogEventView, int, int, error) {
	return deserializeWildcardMatch(deserializer, irBuf, mergedQuery, timeInterval)
}

// fourByteDeserializer contains both a common CLP IR deserializer and stores
// the previously seen log event's timestamp. The previous timestamp is
// necessary to calculate the current timestamp as four byte encoding only
// encodes the timestamp delta between the current log event and the previous.
type fourByteDeserializer struct {
	commonDeserializer
	prevTimestamp ffi.EpochTimeMs
}

// DeserializeLogEvent attempts to read the next log event from the IR stream in
// irBuf, returning the deserialized [ffi.LogEventView], the position read to in
// irBuf (the end of the log event in irBuf), and an error. On error returns:
//   - nil *ffi.LogEventView
//   - 0 position
//   - [IrError] error: CLP failed to successfully deserialize
//   - [EndOfIr] error: CLP found the IR stream EOF tag
func (deserializer *fourByteDeserializer) DeserializeLogEvent(
	irBuf []byte,
) (*ffi.LogEventView, int, error) {
	return deserializeLogEvent(deserializer, irBuf)
}

// DeserializeWildcardMatchWithTimeInterval attempts to read the next log event
// from the IR stream in irBuf that matches mergedQuery within timeInterval. It
// returns the deserialized [ffi.LogEventView], the position read to in irBuf
// (the end of the log event in irBuf), the index of the matched query in
// mergedQuery, and an error. On error returns:
//   - nil *ffi.LogEventView
//   - 0 position
//   - -1 index
//   - [IrError] error: CLP failed to successfully deserialize
//   - [EndOfIr] error: CLP found the IR stream EOF tag
func (deserializer *fourByteDeserializer) DeserializeWildcardMatchWithTimeInterval(
	irBuf []byte,
	mergedQuery search.MergedWildcardQuery,
	timeInterval search.TimestampInterval,
) (*ffi.LogEventView, int, int, error) {
	return deserializeWildcardMatch(deserializer, irBuf, mergedQuery, timeInterval)
}

func deserializeLogEvent(
	deserializer Deserializer,
	irBuf []byte,
) (*ffi.LogEventView, int, error) {
	if 0 >= len(irBuf) {
		return nil, 0, IncompleteIr
	}

	var pos C.size_t
	var event C.LogEventView
	var err error
	switch irs := deserializer.(type) {
	case *eightByteDeserializer:
		err = IrError(C.ir_deserializer_deserialize_eight_byte_log_event(
			newCByteSpan(irBuf),
			irs.cptr,
			&pos,
			&event,
		))
	case *fourByteDeserializer:
		err = IrError(C.ir_deserializer_deserialize_four_byte_log_event(
			newCByteSpan(irBuf),
			irs.cptr,
			&pos,
			&event,
		))
	}
	if Success != err {
		return nil, 0, err
	}

	return &ffi.LogEventView{
			LogMessageView: unsafe.String(
				(*byte)((unsafe.Pointer)(event.m_log_message.m_data)),
				event.m_log_message.m_size,
			),
			Timestamp: ffi.EpochTimeMs(event.m_timestamp),
			UtcOffset: ffi.EpochTimeMs(event.m_utc_offset),
		},
		int(pos),
		nil
}

func deserializeWildcardMatch(
	deserializer Deserializer,
	irBuf []byte,
	mergedQuery search.MergedWildcardQuery,
	time search.TimestampInterval,
) (*ffi.LogEventView, int, int, error) {
	if 0 >= len(irBuf) {
		return nil, 0, -1, IncompleteIr
	}

	var pos C.size_t
	var event C.LogEventView
	var match C.size_t
	var err error
	switch irs := deserializer.(type) {
	case *eightByteDeserializer:
		err = IrError(C.ir_deserializer_deserialize_eight_byte_wildcard_match(
			newCByteSpan(irBuf),
			irs.cptr,
			C.TimestampInterval{C.int64_t(time.Lower), C.int64_t(time.Upper)},
			newMergedWildcardQueryView(mergedQuery),
			&pos,
			&event,
			&match,
		))
	case *fourByteDeserializer:
		err = IrError(C.ir_deserializer_deserialize_four_byte_wildcard_match(
			newCByteSpan(irBuf),
			irs.cptr,
			C.TimestampInterval{C.int64_t(time.Lower), C.int64_t(time.Upper)},
			newMergedWildcardQueryView(mergedQuery),
			&pos,
			&event,
			&match,
		))
	}
	if Success != err {
		return nil, 0, -1, err
	}

	return &ffi.LogEventView{
			LogMessageView: unsafe.String(
				(*byte)((unsafe.Pointer)(event.m_log_message.m_data)),
				event.m_log_message.m_size,
			),
			Timestamp: ffi.EpochTimeMs(event.m_timestamp),
			UtcOffset: ffi.EpochTimeMs(event.m_utc_offset),
		},
		int(pos),
		int(match),
		nil
}
