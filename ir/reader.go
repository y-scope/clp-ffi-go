package ir

import (
	"io"
	"strings"

	"github.com/y-scope/clp-ffi-go/ffi"
	"github.com/y-scope/clp-ffi-go/search"
)

// Reader abstracts maintenance of a buffer containing a [Deserializer]. It
// keeps track of the range [start, end) in the buffer containing valid,
// unconsumed CLP IR. [NewReader] will construct a Reader with the appropriate
// Deserializer based on the consumed CLP IR preamble. The buffer will grow if
// it is not large enough to service a read call (e.g. it cannot hold the next
// log event in the IR). Close must be called to free the underlying memory and
// failure to do so will result in a memory leak.
type Reader struct {
	Deserializer
	ioReader io.Reader
	buf      []byte
	start    int
	end      int
}

// NewReaderSize creates a new [Reader] and uses [DeserializePreamble] to read a
// CLP IR preamble from the [io.Reader], r. size denotes the initial size to use
// for the Reader's buffer that the io.Reader is read into. This buffer will
// grow if it is too small to contain the preamble or next log event. Returns:
//   - success: valid [*Reader], nil
//   - error: nil [*Reader], error propagated from [DeserializePreamble] or
//     [io.Reader.Read]
func NewReaderSize(r io.Reader, size int) (*Reader, error) {
	irr := &Reader{nil, r, make([]byte, size), 0, 0}
	var err error
	if _, err = irr.read(); nil != err {
		return nil, err
	}
	for {
		irr.Deserializer, irr.start, err = DeserializePreamble(irr.buf[irr.start:irr.end])
		if IncompleteIr != err {
			break
		}
		if _, err = irr.fillBuf(); nil != err {
			break
		}
	}
	if nil != err {
		return nil, err
	}
	return irr, nil
}

// Returns [NewReaderSize] with a default buffer size of 1MB.
func NewReader(r io.Reader) (*Reader, error) {
	return NewReaderSize(r, 1024*1024)
}

// Close will delete the underlying C++ allocated memory used by the
// deserializer. Failure to call Close will result in a memory leak.
func (self *Reader) Close() error {
	return self.Deserializer.Close()
}

// Read uses [Deserializer.DeserializeLogEvent] to read from the CLP IR byte stream. The underlying
// buffer will grow if it is too small to contain the next log event. On error returns:
//   - nil *ffi.LogEventView
//   - error propagated from [Deserializer.DeserializeLogEvent] or [io.Reader.Read]
func (self *Reader) Read() (*ffi.LogEventView, error) {
	var event *ffi.LogEventView
	var pos int
	var err error
	for {
		event, pos, err = self.DeserializeLogEvent(self.buf[self.start:self.end])
		if IncompleteIr != err {
			break
		}
		if _, err = self.fillBuf(); nil != err {
			break
		}
	}
	if nil != err {
		return nil, err
	}
	self.start += pos
	return event, nil
}

func (self *Reader) ReadToWildcardMatch(
	timeInterval search.TimestampInterval,
	queries []search.WildcardQuery,
) (*ffi.LogEventView, int, error) {
	var event *ffi.LogEventView
	var pos int
	var matchingQuery int
	var err error
	mergedQuery := search.MergeWildcardQueries(queries)
	for {
		event, pos, matchingQuery, err = self.DeserializeWildcardMatch(
			self.buf[self.start:self.end],
			timeInterval,
			mergedQuery,
		)
		if IncompleteIr != err {
			break
		}
		if _, err = self.fillBuf(); nil != err {
			break
		}
	}
	if nil != err {
		return nil, -1, err
	}
	self.start += pos
	return event, matchingQuery, nil
}

// Read the CLP IR byte stream until f returns true for a [ffi.LogEventView].
// The successful LogEvent is returned. Errors are propagated from [Read].
func (self *Reader) ReadToFunc(
	f func(*ffi.LogEventView) bool,
) (*ffi.LogEventView, error) {
	for {
		event, err := self.Read()
		if nil != err {
			return event, err
		}
		if f(event) {
			return event, nil
		}
	}
}

// Read the CLP IR stream until a [ffi.LogEventView] is greater than or equal to
// the given timestamp. Errors are propagated from [ReadToFunc].
func (self *Reader) ReadToEpochTime(
	time ffi.EpochTimeMs,
) (*ffi.LogEventView, error) {
	return self.ReadToFunc(func(event *ffi.LogEventView) bool { return event.Timestamp >= time })
}

// Read the CLP IR stream until [strings/Contains] returns true for a
// [ffi.LogEventView] and the given sub string. Errors are propagated from
// [ReadToFunc].
func (self *Reader) ReadToContains(substr string) (*ffi.LogEventView, error) {
	fn := func(event *ffi.LogEventView) bool {
		return strings.Contains(event.LogMessageView, substr)
	}
	return self.ReadToFunc(fn)
}

// Read the CLP IR stream until [strings/HasPrefix] returns true for a
// [ffi.LogEventView] and the given prefix. Errors are propagated from
// [ReadToFunc].
func (self *Reader) ReadToPrefix(prefix string) (*ffi.LogEventView, error) {
	fn := func(event *ffi.LogEventView) bool {
		return strings.HasPrefix(event.LogMessageView, prefix)
	}
	return self.ReadToFunc(fn)
}

// Read the CLP IR stream until [strings/HasSuffix] returns true for a
// [ffi.LogEventView] and the given suffix. Errors are propagated from
// [ReadToFunc].
func (self *Reader) ReadToSuffix(suffix string) (*ffi.LogEventView, error) {
	fn := func(event *ffi.LogEventView) bool {
		return strings.HasSuffix(event.LogMessageView, suffix)
	}
	return self.ReadToFunc(fn)
}

// fillBuf shifts the remaining valid IR in [Reader.buf] to the front and then
// calls [io.Reader.Read] to fill the remainder with more IR. Before reading into
// the buffer, it is doubled if more than half of it is unconsumed IR.
// Forwards the return of [io.Reader.Read].
func (self *Reader) fillBuf() (int, error) {
	if (self.end - self.start) > len(self.buf)/2 {
		buf := make([]byte, len(self.buf)*2)
		copy(buf, self.buf[self.start:self.end])
		self.buf = buf
	} else {
		copy(self.buf, self.buf[self.start:self.end])
	}
	self.end -= self.start
	self.start = 0
	n, err := self.read()
	return n, err
}

// read is a wrapper around a io.Reader.Read call. It uses the correct range in
// buf and adjusts the range accordingly. Always returns the number of bytes
// read. On success nil is returned. On failure an error is forwarded from
// [io.Reader], unless n > 0 and io.EOF == err as we have not yet consumed the
// CLP IR.
func (self *Reader) read() (int, error) {
	n, err := self.ioReader.Read(self.buf[self.end:])
	self.end += n
	if nil != err && io.EOF != err {
		return n, err
	}
	return n, nil
}
