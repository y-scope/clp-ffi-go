package ir

/*
#include <ffi_go/ir/encoder.h>
*/
import "C"

import (
	"unsafe"

	"github.com/y-scope/clp-ffi-go/ffi"
)

// An Encoder takes logging objects (commonly used/created by logging libraries)
// and encodes them as CLP IR. Close must be called to free the underlying
// memory and failure to do so will result in a memory leak.
type Encoder[T EightByteEncoding | FourByteEncoding] interface {
	EncodeLogMessage(logMessage ffi.LogMessage) (*LogMessageView[T], error)
	Close() error
}

// Return a new Encoder that produces IR using [EightByteEncoding].
func EightByteEncoder() (Encoder[EightByteEncoding], error) {
	return &eightByteEncoder{C.ir_encoder_eight_byte_new()}, nil
}

// Return a new Encoder that produces IR using [FourByteEncoding].
func FourByteEncoder() (Encoder[FourByteEncoding], error) {
	return &fourByteEncoder{C.ir_encoder_four_byte_new()}, nil
}

type eightByteEncoder struct {
	cptr unsafe.Pointer
}

// Close will delete the underlying C++ allocated memory used by the
// deserializer. Failure to call Close will result in a memory leak.
func (self *eightByteEncoder) Close() error {
	if nil != self.cptr {
		C.ir_encoder_eight_byte_close(self.cptr)
		self.cptr = nil
	}
	return nil
}

// Encode a log message into CLP IR, returning a view of the encoded message.
func (self *eightByteEncoder) EncodeLogMessage(
	logMessage ffi.LogMessage,
) (*LogMessageView[EightByteEncoding], error) {
	var logtype C.StringView
	var vars C.Int64tSpan
	var dictVars C.StringView
	var dictVarEndOffsets C.Int32tSpan
	err := IrError(C.ir_encoder_encode_eight_byte_log_message(
		newCStringView(logMessage),
		self.cptr,
		&logtype,
		&vars,
		&dictVars,
		&dictVarEndOffsets,
	))
	if Success != err {
		return nil, EncodeError
	}
	return newLogMessageView[EightByteEncoding](logtype, vars, dictVars, dictVarEndOffsets), nil
}

type fourByteEncoder struct {
	cptr unsafe.Pointer
}

// Close will delete the underlying C++ allocated memory used by the
// deserializer. Failure to call Close will result in a memory leak.
func (self *fourByteEncoder) Close() error {
	if nil != self.cptr {
		C.ir_encoder_four_byte_close(self.cptr)
		self.cptr = nil
	}
	return nil
}

// Encode a log message into CLP IR, returning a view of the encoded message.
func (self *fourByteEncoder) EncodeLogMessage(
	logMessage ffi.LogMessage,
) (*LogMessageView[FourByteEncoding], error) {
	var logtype C.StringView
	var vars C.Int32tSpan
	var dictVars C.StringView
	var dictVarEndOffsets C.Int32tSpan
	err := IrError(C.ir_encoder_encode_four_byte_log_message(
		newCStringView(logMessage),
		self.cptr,
		&logtype,
		&vars,
		&dictVars,
		&dictVarEndOffsets,
	))
	if Success != err {
		return nil, EncodeError
	}
	return newLogMessageView[FourByteEncoding](logtype, vars, dictVars, dictVarEndOffsets), nil
}
