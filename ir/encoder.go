package ir

/*
#include "../cpp/src/ir/encoding.h"
*/
import "C"

import (
	"runtime"
	"unsafe"

	"github.com/y-scope/clp-ffi-go" // ffi::types + necessary for linkage
)

type IrEncoder interface {
	EncodeMessage(ts ffi.EpochTimeMs, msg string) ([]byte, int)
	EncodeMessageUnsafe(ts ffi.EpochTimeMs, msg string) ([]byte, int)
	TimestampInfo() TimestampInfo
}

func EightByteEncodePreamble(
	ts_pattern string,
	ts_pattern_syntax string,
	time_zone_id string,
) (EightByteIrStream, []byte, int) {
	irs, preamble, ret := EightByteEncodePreambleUnsafe(ts_pattern, ts_pattern_syntax,
		time_zone_id)
	if 0 != ret {
		return irs, nil, ret
	}
	safePreamble := make([]byte, len(preamble))
	copy(safePreamble, preamble)
	return irs, safePreamble, 0
}

func EightByteEncodePreambleUnsafe(
	ts_pattern string,
	ts_pattern_syntax string,
	time_zone_id string,
) (EightByteIrStream, []byte, int) {
	var bufPtr unsafe.Pointer
	var bufSize uint64
	irs := EightByteIrStream{
		irStream[EightByteEncodedVariable]{
			TimestampInfo{ts_pattern, ts_pattern_syntax, time_zone_id}, nil,
		},
	}
	irs.cPtr = C.eight_byte_encode_preamble(
		unsafe.Pointer(&[]byte(ts_pattern)[0]), C.size_t(len(ts_pattern)),
		unsafe.Pointer(&[]byte(ts_pattern_syntax)[0]), C.size_t(len(ts_pattern_syntax)),
		unsafe.Pointer(&[]byte(time_zone_id)[0]), C.size_t(len(time_zone_id)),
		&bufPtr, unsafe.Pointer(&bufSize))
	buf := unsafe.Slice((*byte)(bufPtr), bufSize)
	if nil == buf {
		return irs, nil, -2
	}
	runtime.SetFinalizer(&irs,
		func(irs *EightByteIrStream) { C.delete_ir_stream_state(irs.cPtr) })
	return irs, buf, 0
}

func FourByteEncodePreamble(
	ts_pattern string,
	ts_pattern_syntax string,
	time_zone_id string,
	reference_ts ffi.EpochTimeMs,
) (FourByteIrStream, []byte, int) {
	irs, preamble, ret := FourByteEncodePreambleUnsafe(ts_pattern, ts_pattern_syntax,
		time_zone_id, reference_ts)
	if 0 != ret {
		return irs, nil, ret
	}
	safePreamble := make([]byte, len(preamble))
	copy(safePreamble, preamble)
	return irs, safePreamble, 0
}

func FourByteEncodePreambleUnsafe(
	ts_pattern string,
	ts_pattern_syntax string,
	time_zone_id string,
	reference_ts ffi.EpochTimeMs,
) (FourByteIrStream, []byte, int) {
	var bufPtr unsafe.Pointer
	var bufSize uint64
	irs := FourByteIrStream{
		irStream[FourByteEncodedVariable]{
			TimestampInfo{ts_pattern, ts_pattern_syntax, time_zone_id}, nil,
		},
		reference_ts,
	}
	irs.cPtr = C.four_byte_encode_preamble(
		unsafe.Pointer(&[]byte(ts_pattern)[0]), C.size_t(len(ts_pattern)),
		unsafe.Pointer(&[]byte(ts_pattern_syntax)[0]), C.size_t(len(ts_pattern_syntax)),
		unsafe.Pointer(&[]byte(time_zone_id)[0]), C.size_t(len(time_zone_id)),
		C.int64_t(reference_ts), &bufPtr, unsafe.Pointer(&bufSize))
	buf := unsafe.Slice((*byte)(bufPtr), bufSize)
	if nil == buf {
		return irs, nil, -2
	}
	runtime.SetFinalizer(&irs,
		func(irs *FourByteIrStream) { C.delete_ir_stream_state(irs.cPtr) })
	return irs, buf, 0
}

func (self *EightByteIrStream) EncodeMessage(ts ffi.EpochTimeMs, msg string) ([]byte, int) {
	return encodeMessage(self, ts, msg)
}

func (self *FourByteIrStream) EncodeMessage(ts ffi.EpochTimeMs, msg string) ([]byte, int) {
	return encodeMessage(self, ts, msg)
}

func encodeMessage(irEncoder IrEncoder, ts ffi.EpochTimeMs, msg string) ([]byte, int) {
	buf, ret := irEncoder.EncodeMessageUnsafe(ts, msg)
	if 0 != ret {
		return nil, ret
	}
	safeBuf := make([]byte, len(buf))
	copy(safeBuf, buf)
	return safeBuf, 0
}

func (self *EightByteIrStream) EncodeMessageUnsafe(ts ffi.EpochTimeMs, msg string) ([]byte, int) {
	return encodeMessageUnsafe(self, ts, msg)
}

func (self *FourByteIrStream) EncodeMessageUnsafe(ts ffi.EpochTimeMs, msg string) ([]byte, int) {
	buf, ret := encodeMessageUnsafe(self, self.prevTimestamp-ts, msg)
	if 0 != ret {
		return nil, ret
	}
	self.prevTimestamp = ts
	return buf, ret
}

// returns 0 on success, >0 on error, <0 on c error
// returned byte slice points to c memory and is only valid until the next call
// to encodeMessage (from either EncodeMessage or EncodeMessageUnsafe)
func encodeMessageUnsafe[T EightByteIrStream | FourByteIrStream](
	irstream *T,
	timestampOrDelta ffi.EpochTimeMs,
	msg string,
) ([]byte, int) {
	var ret C.int
	var bufPtr unsafe.Pointer
	var bufSize uint64

	switch irs := any(irstream).(type) {
	case *EightByteIrStream:
		ret = C.eight_byte_encode_message(irs.cPtr, C.int64_t(timestampOrDelta),
			unsafe.Pointer(&[]byte(msg)[0]), C.size_t(len(msg)),
			&bufPtr, unsafe.Pointer(&bufSize))
	case *FourByteIrStream:
		ret = C.four_byte_encode_message(irs.cPtr, C.int64_t(timestampOrDelta),
			unsafe.Pointer(&[]byte(msg)[0]), C.size_t(len(msg)),
			&bufPtr, unsafe.Pointer(&bufSize))
	default:
		return nil, 2
	}
	if 0 > ret {
		return nil, int(ret)
	}
	buf := unsafe.Slice((*byte)(bufPtr), bufSize)
	if nil == buf {
		return nil, 3
	}
	return buf, 0
}
