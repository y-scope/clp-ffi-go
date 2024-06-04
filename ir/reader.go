package ir

import (
	"io"
	"math"
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
func (reader *Reader) Close() error {
	return reader.Deserializer.Close()
}

// Read uses [Deserializer].DeserializeLogEvent to read from the CLP IR byte stream. The
// underlying buffer will grow if it is too small to contain the next log event. On error returns:
//   - nil [*ffi.LogEventView]
//   - error propagated from [Deserializer].DeserializeLogEvent or [io.Reader.Read]
func (reader *Reader) Read() (*ffi.LogEventView, error) {
	var event *ffi.LogEventView
	var pos int
	var err error
	for {
		event, pos, err = reader.DeserializeLogEvent(reader.buf[reader.start:reader.end])
		if IncompleteIr != err {
			break
		}
		if _, err = reader.fillBuf(); nil != err {
			break
		}
	}
	if nil != err {
		return nil, err
	}
	reader.start += pos
	return event, nil
}

// ReadToWildcardMatch wraps ReadToWildcardMatchWithTimeInterval, attempting to
// read the next log event that matches any query in queries, within the entire
// IR. It forwards the result of ReadToWildcardMatchWithTimeInterval.
func (reader *Reader) ReadToWildcardMatch(
	queries []search.WildcardQuery,
) (*ffi.LogEventView, int, error) {
	return reader.ReadToWildcardMatchWithTimeInterval(
		queries,
		search.TimestampInterval{Lower: 0, Upper: math.MaxInt64},
	)
}

// ReadToWildcardMatchWithTimeInterval attempts to read the next log event that
// matches any query in queries, within timeInterval. It returns the
// deserialized [ffi.LogEventView], the index of the matched query in queries,
// and an error. On error returns:
//   - nil *ffi.LogEventView
//   - -1 index
//   - [IrError] error: CLP failed to successfully deserialize
//   - [EndOfIr] error: CLP found the IR stream EOF tag
func (reader *Reader) ReadToWildcardMatchWithTimeInterval(
	queries []search.WildcardQuery,
	timeInterval search.TimestampInterval,
) (*ffi.LogEventView, int, error) {
	var event *ffi.LogEventView
	var pos int
	var matchingQuery int
	var err error
	mergedQuery := search.MergeWildcardQueries(queries)
	for {
		event, pos, matchingQuery, err = reader.DeserializeWildcardMatchWithTimeInterval(
			reader.buf[reader.start:reader.end],
			mergedQuery,
			timeInterval,
		)
		if IncompleteIr != err {
			break
		}
		if _, err = reader.fillBuf(); nil != err {
			break
		}
	}
	if nil != err {
		return nil, -1, err
	}
	reader.start += pos
	return event, matchingQuery, nil
}

// Read the CLP IR byte stream until f returns true for a [ffi.LogEventView].
// The successful LogEvent is returned. Errors are propagated from [Reader.Read].
func (reader *Reader) ReadToFunc(
	f func(*ffi.LogEventView) bool,
) (*ffi.LogEventView, error) {
	for {
		event, err := reader.Read()
		if nil != err {
			return event, err
		}
		if f(event) {
			return event, nil
		}
	}
}

// Read the CLP IR stream until a [ffi.LogEventView] is greater than or equal to
// the given timestamp. Errors are propagated from [Reader.ReadToFunc].
func (reader *Reader) ReadToEpochTime(
	time ffi.EpochTimeMs,
) (*ffi.LogEventView, error) {
	return reader.ReadToFunc(func(event *ffi.LogEventView) bool { return event.Timestamp >= time })
}

// Read the CLP IR stream until [strings/Contains] returns true for a
// [ffi.LogEventView] and the given sub string. Errors are propagated from
// [Reader.ReadToFunc].
func (reader *Reader) ReadToContains(substr string) (*ffi.LogEventView, error) {
	fn := func(event *ffi.LogEventView) bool {
		return strings.Contains(event.LogMessageView, substr)
	}
	return reader.ReadToFunc(fn)
}

// Read the CLP IR stream until [strings/HasPrefix] returns true for a
// [ffi.LogEventView] and the given prefix. Errors are propagated from
// [Reader.ReadToFunc].
func (reader *Reader) ReadToPrefix(prefix string) (*ffi.LogEventView, error) {
	fn := func(event *ffi.LogEventView) bool {
		return strings.HasPrefix(event.LogMessageView, prefix)
	}
	return reader.ReadToFunc(fn)
}

// Read the CLP IR stream until [strings/HasSuffix] returns true for a
// [ffi.LogEventView] and the given suffix. Errors are propagated from
// [Reader.ReadToFunc].
func (reader *Reader) ReadToSuffix(suffix string) (*ffi.LogEventView, error) {
	fn := func(event *ffi.LogEventView) bool {
		return strings.HasSuffix(event.LogMessageView, suffix)
	}
	return reader.ReadToFunc(fn)
}

// fillBuf shifts the remaining valid IR in [Reader.buf] to the front and then
// calls [io.Reader.Read] to fill the remainder with more IR. Before reading into
// the buffer, it is doubled if more than half of it is unconsumed IR.
// Forwards the return of [io.Reader.Read].
func (reader *Reader) fillBuf() (int, error) {
	if (reader.end - reader.start) > len(reader.buf)/2 {
		buf := make([]byte, len(reader.buf)*2)
		copy(buf, reader.buf[reader.start:reader.end])
		reader.buf = buf
	} else {
		copy(reader.buf, reader.buf[reader.start:reader.end])
	}
	reader.end -= reader.start
	reader.start = 0
	n, err := reader.read()
	return n, err
}

// read is a wrapper around a io.Reader.Read call. It uses the correct range in
// buf and adjusts the range accordingly. Always returns the number of bytes
// read. On success nil is returned. On failure an error is forwarded from
// [io.Reader], unless n > 0 and io.EOF == err as we have not yet consumed the
// CLP IR.
func (reader *Reader) read() (int, error) {
	n, err := reader.ioReader.Read(reader.buf[reader.end:])
	reader.end += n
	if nil != err && io.EOF != err {
		return n, err
	}
	return n, nil
}
