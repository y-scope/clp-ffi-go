package ir

import (
	"bytes"
	"io"

	"github.com/y-scope/clp-ffi-go" // ffi::types + necessary for linkage
)

// IrReader abstracts maintenance of a buffer containing an IR stream. It keeps
// track of the range in the buffer containing valid, unconsumed IR. It does
// not store a Reader to allow callers to mutate the Reader as necessary.
type IrReader struct {
	IrDecoder
	buf   []byte
	start int
	end   int
}

// adjust_buf mutates the IrReader.buf so that the next read call has space to
// fill. If the start of IrReader.buf is not 0 the contents of buf will be
// shifted back, so that end -= start and start = 0. If start is already 0 the
// buffer is grown.
func (self *IrReader) adjust_buf() int {
	if 0 == self.start {
		buf := make([]byte, len(self.buf)*2)
		copy(buf, self.buf[self.start:self.end])
		self.buf = buf
	} else {
		copy(self.buf, self.buf[self.start:self.end])
	}
	self.end -= self.start
	self.start = 0
	return len(self.buf)
}

// read is a wrapper around the Read call to the io.Reader. It uses the correct
// range in buf and adjusts the range accordingly. On success nil is returned.
// On failure an error whose type depends on the io.Reader is returned.
// Note we do not return io.EOF if n > 0 as we have not yet consumed the IR.
func (self *IrReader) read(r io.Reader) error {
	n, err := r.Read(self.buf[self.end:])
	if nil != err && io.EOF != err {
		return err
	}
	self.end += n
	return nil
}

// ReadPreamble uses [DecodePreamble] to read an IR stream preamble from r.
// bufSize denotes the initial size to use for the underlying buffer io.Reader
// is read into. This buffer will grow if it is too small to contain the
// preamble or next log event.
// Return values:
//   - nil == error: success
//   - nil != error:
//     - type [IRError] or [encoding/json]: error propagated from [DecodePreamble]
//     - type from io.Reader: io.Reader.Read failed
func ReadPreamble(r io.Reader, bufSize int) (IrReader, error) {
	irr := IrReader{nil, make([]byte, bufSize), 0, 0}

	if err := irr.read(r); nil != err {
		return irr, err
	}

	for {
		var err error
		irr.IrDecoder, irr.start, err = DecodePreamble(irr.buf[irr.start:irr.end])
		if nil == err {
			return irr, nil
		} else if IncompleteIR == err {
			irr.adjust_buf()
			if err := irr.read(r); nil != err {
				return irr, err
			}
		} else {
			return irr, err
		}
	}
}


// ReadNextLogEvent uses [DecodeNextLogEvent] to read from the IR stream in r.
// bufSize denotes the initial size to use for the underlying buffer io.Reader
// is read into. This buffer will grow if it is too small to contain the
// preamble or next log event.
// Return values:
//   - nil == error: success
//   - IRError.Eof: CLP found the IR stream EOF tag
//   - io.EOF: io.Reader.Read got EOF
//   - else:
//     - type [IRError]: error propagated from [DecodeNextLogEvent]
//     - type from io.Reader: io.Reader.Read failed
func (self *IrReader) ReadNextLogEvent(r io.Reader) (ffi.LogEvent, error) {
	for {
		event, offset, err := self.DecodeNextLogEvent(self.buf[self.start:self.end])
		if nil == err {
			self.start += offset
			return event, nil
		} else if IncompleteIR == err {
			self.adjust_buf()
			if err := self.read(r); nil != err {
				return event, err
			}
		} else {
			return event, err
		}
	}
}

// Read the IR stream using the io.Reader until f returns true for a
// [ffi.LogEvent]. The succeeding LogEvent is returned. Errors are propagated
// from ReadNextLogEvent.
func (self *IrReader) ReadToFunc(r io.Reader, f func(ffi.LogEvent) bool) (ffi.LogEvent, error) {
	for {
		event, err := self.ReadNextLogEvent(r)
		if nil != err {
			return event, err
		}
		if f(event) {
			return event, nil
		}
	}
}

// Read the IR stream using the io.Reader until [ffi.LogEvent.Timestamp] >=
// time. Errors are propagated from ReadNextLogEvent.
func (self *IrReader) ReadToEpochTime(r io.Reader, time ffi.EpochTimeMs) (ffi.LogEvent, error) {
	return self.ReadToFunc(r, func(e ffi.LogEvent) bool { return e.Timestamp >= time })
}

// Read the IR stream using the io.Reader until [bytes/Contains] returns true
// for [ffi.LogEvent.Msg] and subslice. Errors are propagated from ReadNextLogEvent.
func (self *IrReader) ReadToContains(r io.Reader, subslice []byte) (ffi.LogEvent, error) {
	return self.ReadToFunc(r, func(e ffi.LogEvent) bool { return bytes.Contains(e.Msg, subslice) })
}

// Read the IR stream using the io.Reader until [bytes/HasPrefix] returns true
// for [ffi.LogEvent.Msg] and prefix. Errors are propagated from ReadNextLogEvent.
func (self *IrReader) ReadToPrefix(r io.Reader, prefix []byte) (ffi.LogEvent, error) {
	return self.ReadToFunc(r, func(e ffi.LogEvent) bool { return bytes.HasPrefix(e.Msg, prefix) })
}

// Read the IR stream using the io.Reader until [bytes/HasSuffix] returns true
// for [ffi.LogEvent.Msg] field and suffix. Errors are propagated from ReadNextLogEvent.
func (self *IrReader) ReadToSuffix(r io.Reader, suffix []byte) (ffi.LogEvent, error) {
	return self.ReadToFunc(r, func(e ffi.LogEvent) bool { return bytes.HasSuffix(e.Msg, suffix) })
}
