package ir

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/y-scope/clp-ffi-go/ffi"
)

// Writer builds up a buffer of serialized CLP IR using a [Serializer].
// [NewWriter] will construct a Writer with the appropriate Serializer based on
// the arguments used. Close must be called to free the underlying memory and
// failure to do so will result in a memory leak. To write a complete IR stream
// Close must be called before the final WriteTo call.
type Writer struct {
	Serializer
	buf bytes.Buffer
}

// Returns [NewWriterSize] with a FourByteEncoding Serializer using the local
// time zone, and a buffer size of 1MB.
func NewWriter() (*Writer, error) {
	return NewWriterSize[FourByteEncoding](1024*1024, time.Local.String())
}

// NewWriterSize creates a new [Writer] with a [Serializer] based on T, and
// writes a CLP IR preamble. The preamble is stored inside the Writer's internal
// buffer to be written out later. The size parameter denotes the initial buffer
// size to use and timeZoneId denotes the time zone of the source producing the
// log events, so that local times (any time that is not a unix timestamp) are
// handled correctly.
//   - success: valid [*Writer], nil
//   - error: nil [*Writer], invalid type error or an error propagated from
//     [FourByteSerializer], [EightByteSerializer], or [bytes.Buffer.Write]
func NewWriterSize[T EightByteEncoding | FourByteEncoding](
	size int,
	timeZoneId string,
) (*Writer, error) {
	var irw Writer
	irw.buf.Grow(size)

	var irView BufView
	var err error
	var t T
	switch any(t).(type) {
	case EightByteEncoding:
		irw.Serializer, irView, err = EightByteSerializer(
			"",
			"",
			timeZoneId,
		)
	case FourByteEncoding:
		irw.Serializer, irView, err = FourByteSerializer(
			"",
			"",
			timeZoneId,
			ffi.EpochTimeMs(time.Now().UnixMilli()),
		)
	default:
		err = fmt.Errorf("Invalid type: %T", t)
	}
	if nil != err {
		return nil, err
	}
	_, err = irw.buf.Write(irView)
	if nil != err {
		return nil, err
	}
	return &irw, nil
}

// Close will write a null byte denoting the end of the IR stream and delete the
// underlying C++ allocated memory used by the serializer. Failure to call Close
// will result in a memory leak.
func (self *Writer) Close() error {
	self.buf.WriteByte(0x0)
	return self.Serializer.Close()
}

// CloseTo is a combination of [Close] and [WriteTo]. It will completely close
// the Writer (and underlying serializer) and write the data out to the
// io.Writer.
// Returns:
//   - success: number of bytes written, nil
//   - error: number of bytes written, error propagated from [WriteTo]
func (self *Writer) CloseTo(w io.Writer) (int64, error) {
	self.Close()
	return self.WriteTo(w)
}

// Bytes returns a slice of the Writer's internal buffer. The slice is valid for
// use only until the next buffer modification (that is, only until the next
// call to Write, WriteTo, or Reset).
func (self *Writer) Bytes() []byte {
	return self.buf.Bytes()
}

// Reset resets the buffer to be empty, but it retains the underlying storage
// for use by future writes.
func (self *Writer) Reset() {
	self.buf.Reset()
}

// Write uses [SerializeLogEvent] to serialize the provided log event to CLP IR
// and then stores it in the internal buffer. Returns:
//   - success: number of bytes written, nil
//   - error: number of bytes written (can be 0), error propagated from
//     [SerializeLogEvent] or [bytes.Buffer.Write]
func (self *Writer) Write(event ffi.LogEvent) (int, error) {
	irView, err := self.SerializeLogEvent(event)
	if nil != err {
		return 0, err
	}
	// bytes.Buffer.Write will always return nil for err (https://pkg.go.dev/bytes#Buffer.Write)
	// However, err is still propagated to correctly alert the user in case this ever changes. If
	// Write can fail in the future, we should either:
	//   1. fix the issue and retry the write
	//   2. store irView and provide a retry API (allowing the user to fix the issue and retry)
	n, err := self.buf.Write(irView)
	if nil != err {
		return n, err
	}
	return n, nil
}

// WriteTo writes data to w until the buffer is drained or an error occurs. If
// no error occurs the buffer is reset. On an error the user is expected to use
// [self.Bytes] and [self.Reset] to manually handle the buffer's contents before
// continuing. Returns:
//   - success: number of bytes written, nil
//   - error: number of bytes written, error propagated from
//     [bytes.Buffer.WriteTo]
func (self *Writer) WriteTo(w io.Writer) (int64, error) {
	n, err := self.buf.WriteTo(w)
	if nil == err {
		self.buf.Reset()
	}
	return n, err
}
