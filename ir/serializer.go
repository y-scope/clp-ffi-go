package ir

/*
#include <ffi_go/defs.h>
#include <ffi_go/ir/serializer.h>
*/
import "C"

import (
	"unsafe"

	"github.com/y-scope/clp-ffi-go/ffi"
	"github.com/vmihailenco/msgpack/v5"
)

// A Serializer exports functions to serialize log events into a CLP IR byte
// stream. Serialization functions only return views (slices) of IR bytes,
// leaving their use to the user. Each Serializer owns its own unique underlying
// memory for the views it produces/returns. This memory is reused for each
// view, so to persist the contents the memory must be copied into another
// object. Close must be called to free the underlying memory and failure to do
// so will result in a memory leak.
type Serializer interface {
	SerializeLogEvent(logEvent ffi.LogEvent) (BufView, error)
	SerializeMsgPackBytes(msgPackBytes []byte) (BufView, error)
	Close() error
}

// EightByteSerializer creates and returns a new Serializer that writes eight
// byte encoded CLP IR and serializes a IR preamble into a BufView using it. On
// error returns:
//   - nil Serializer
//   - nil BufView
//   - [IrError] error: CLP failed to successfully serialize
func EightByteSerializer() (Serializer, BufView, error) {
	var irView C.ByteSpan
	irs := eightByteSerializer{commonSerializer{nil}}
	if err := IrError(C.ir_serializer_eight_byte_create(
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
func FourByteSerializer() (Serializer, BufView, error) {
	var irView C.ByteSpan
	irs := fourByteSerializer{commonSerializer{nil}}
	if err := IrError(C.ir_serializer_four_byte_create(
		&irs.cptr,
		&irView,
	)); Success != err {
		return nil, nil, err
	}
	return &irs, unsafe.Slice((*byte)(irView.m_data), irView.m_size), nil
}

// commonSerializer contains fields common to all types of CLP IR serializers.
// cptr holds a reference to the underlying C++ objected used as backing storage
// for the Views returned by the serializer. Close must be called to free this
// underlying memory and failure to do so will result in a memory leak.
type commonSerializer struct {
	cptr unsafe.Pointer
}

// Create a distinct type so we know the type of the underlying serializer, but allows the use of
// the same methods.
type eightByteSerializer struct {
	commonSerializer
}

// Closes the serializer by releasing the underlying C++ allocated memory.
// Failure to call Close will result in a memory leak.
func (serializer *eightByteSerializer) Close() error {
	if nil != serializer.cptr {
		C.ir_serializer_eight_byte_close(serializer.cptr)
		serializer.cptr = nil
	}
	return nil
}

// SerializeLogEvent attempts to serialize the log event, into an eight byte encoded CLP IR byte stream.
// On error returns:
//   - a nil BufView
//   - [IrError] based on the failure of the Cgo call
func (serializer *eightByteSerializer) SerializeLogEvent(
	LogEvent ffi.LogEvent,
) (BufView, error) {
	return serializeLogEvent(serializer, LogEvent)
}

// SerializeMsgPackBytes attempts to serialize the log event, event, into a eight
// byte encoded CLP IR byte stream. On error returns:
//   - nil BufView
//   - [IrError] based on the failure of the Cgo call
func (serializer *eightByteSerializer) SerializeMsgPackBytes(
	msgPackBytes []byte,
) (BufView, error) {
	return serializeMsgPackBytes(serializer, msgPackBytes)
}

// Create a distinct type so we know the type of the underlying serializer, but allows the use of
// the same methods.
type fourByteSerializer struct {
	commonSerializer
}

// Closes the serializer by releasing the underlying C++ allocated memory.
// Failure to call Close will result in a memory leak.
func (serializer *fourByteSerializer) Close() error {
	if nil != serializer.cptr {
		C.ir_serializer_four_byte_close(serializer.cptr)
		serializer.cptr = nil
	}
	return nil
}

// SerializeLogEvent attempts to serialize the log event, event, into a four
// byte encoded CLP IR byte stream. On error returns:
//   - nil BufView
//   - [IrError] based on the failure of the Cgo call
func (serializer *fourByteSerializer) SerializeLogEvent(
	logEvent ffi.LogEvent,
) (BufView, error) {
	return serializeLogEvent(serializer, logEvent)
}

// SerializeMsgPackBytes attempts to serialize the log event, event, into a four
// byte encoded CLP IR byte stream. On error returns:
//   - nil BufView
//   - [IrError] based on the failure of the Cgo call
func (serializer *fourByteSerializer) SerializeMsgPackBytes(
	msgPackBytes []byte,
) (BufView, error) {
	return serializeMsgPackBytes(serializer, msgPackBytes)
}

func serializeLogEvent(
	serializer Serializer,
	logEvent ffi.LogEvent,
) (BufView, error) {
	msgPackBytes, err := msgpack.Marshal(&logEvent)
	if err != nil {
		return nil, err
	}
	return serializeMsgPackBytes(serializer, msgPackBytes)
}

func serializeMsgPackBytes(
	serializer Serializer,
	msgPackBytes []byte,
) (BufView, error) {
	var irView C.ByteSpan
	var err error

	switch irs := serializer.(type) {
	case *eightByteSerializer:
		err = IrError(C.ir_serializer_eight_byte_serialize_log_event(
			irs.cptr,
			newCByteSpan(msgPackBytes),
			&irView,
		))
	case *fourByteSerializer:
		err = IrError(C.ir_serializer_four_byte_serialize_log_event(
			irs.cptr,
			newCByteSpan(msgPackBytes),
			&irView,
		))
	}
	if Success != err {
		return nil, err
	}
	return unsafe.Slice((*byte)(irView.m_data), irView.m_size), nil
}
