package ir

/*
#include <ffi_go/ir/decoder.h>
*/
import "C"

import (
	"unsafe"

	"github.com/y-scope/clp-ffi-go/ffi"
)

// A Decoder takes objects encoded in CLP IR as input and returns them in their
// natural state prior to encoding. Close must be called to free the underlying
// memory and failure to do so will result in a memory leak.
type Decoder[T EightByteEncoding | FourByteEncoding] interface {
	DecodeLogMessage(irMessage LogMessage[T]) (*ffi.LogMessageView, error)
	Close() error
}

// Return a new Decoder for IR using [EightByteEncoding].
func EightByteDecoder() (Decoder[EightByteEncoding], error) {
	return &eightByteDecoder{commonDecoder{C.ir_decoder_new()}}, nil
}

// Return a new Decoder for IR using [FourByteEncoding].
func FourByteDecoder() (Decoder[FourByteEncoding], error) {
	return &fourByteDecoder{commonDecoder{C.ir_decoder_new()}}, nil
}

type commonDecoder struct {
	cptr unsafe.Pointer
}

// Close will delete the underlying C++ allocated memory used by the
// deserializer. Failure to call Close will result in a memory leak.
func (decoder *commonDecoder) Close() error {
	if nil != decoder.cptr {
		C.ir_decoder_close(decoder.cptr)
		decoder.cptr = nil
	}
	return nil
}

type eightByteDecoder struct {
	commonDecoder
}

// Decode an IR encoded log message, returning a view of the original
// (non-encoded) log message.
func (decoder *eightByteDecoder) DecodeLogMessage(
	irMessage LogMessage[EightByteEncoding],
) (*ffi.LogMessageView, error) {
	var msg C.StringView
	err := IrError(C.ir_decoder_decode_eight_byte_log_message(
		newCStringView(irMessage.Logtype),
		newCInt64tSpan(irMessage.Vars),
		newCStringView(irMessage.DictVars),
		newCInt32tSpan(irMessage.DictVarEndOffsets),
		decoder.cptr,
		&msg,
	))
	if Success != err {
		return nil, DecodeError
	}
	view := unsafe.String((*byte)(unsafe.Pointer(msg.m_data)), msg.m_size)
	return &view, nil
}

type fourByteDecoder struct {
	commonDecoder
}

// Decode an IR encoded log message, returning a view of the original
// (non-encoded) log message.
func (decoder *fourByteDecoder) DecodeLogMessage(
	irMessage LogMessage[FourByteEncoding],
) (*ffi.LogMessageView, error) {
	var msg C.StringView
	err := IrError(C.ir_decoder_decode_four_byte_log_message(
		newCStringView(irMessage.Logtype),
		newCInt32tSpan(irMessage.Vars),
		newCStringView(irMessage.DictVars),
		newCInt32tSpan(irMessage.DictVarEndOffsets),
		decoder.cptr,
		&msg,
	))
	if Success != err {
		return nil, DecodeError
	}
	view := unsafe.String((*byte)(unsafe.Pointer(msg.m_data)), msg.m_size)
	return &view, nil
}
