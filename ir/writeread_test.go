package ir

import (
	"io"
	"testing"
	"time"

	"github.com/y-scope/clp-ffi-go/ffi"
)

func testWriteReadLogMessages(
	t *testing.T,
	args testArgs,
	messages []ffi.LogMessage,
) {
	ioWriter := openIoWriter(t, args)
	irWriter := openIrWriter(t, args, ioWriter)

	utcOffsetToronto := ffi.EpochTimeMs(-4 * 60 * 60 * 1000)
	utcOffsetTokyo := ffi.EpochTimeMs(9 * 60 * 60 * 100)

	_, err := irWriter.WriteUtcOffsetChange(utcOffsetTokyo)
	if nil != err {
		t.Fatalf("ir.Writer.WriteUtcOffsetChange failed: %v", err)
	}

	// Overwrite the previous UTC offset
	_, err = irWriter.WriteUtcOffsetChange(utcOffsetToronto)
	if nil != err {
		t.Fatalf("ir.Writer.WriteUtcOffsetChange failed: %v", err)
	}

	// Serialize the same UTC offset again
	_, err = irWriter.WriteUtcOffsetChange(utcOffsetToronto)
	if nil != err {
		t.Fatalf("ir.Writer.WriteUtcOffsetChange failed: %v", err)
	}

	var events []ffi.LogEvent
	for _, msg := range messages {
		event := ffi.LogEvent{
			LogMessage: msg,
			Timestamp:  ffi.EpochTimeMs(time.Now().UnixMilli()),
		}
		_, err := irWriter.WriteLogEvent(event)
		if nil != err {
			t.Fatalf("ir.Writer.WriteLogEvent failed: %v", err)
		}
		events = append(events, event)
	}
	_, err = irWriter.CloseTo(ioWriter)
	if nil != err {
		t.Fatalf("ir.Writer.CloseTo failed: %v", err)
	}
	ioWriter.Close()

	ioReader := openIoReader(t, args)
	defer ioReader.Close()
	irReader, err := NewReader(ioReader)
	if nil != err {
		t.Fatalf("NewReader failed: %v", err)
	}
	defer irReader.Close()

	for _, event := range events {
		assertIrLogEvent(t, ioReader, irReader, event, utcOffsetToronto)
	}
	assertEndOfIr(t, ioReader, irReader)
}

func openIrWriter(
	t *testing.T,
	args testArgs,
	writer io.Writer,
) *Writer {
	var irWriter *Writer
	var err error
	switch args.encoding {
	case eightByteEncoding:
		irWriter, err = NewWriterSize[EightByteEncoding](1024 * 1024)
	case fourByteEncoding:
		irWriter, err = NewWriterSize[FourByteEncoding](1024 * 1024)
	default:
		t.Fatalf("unsupported encoding: %v", args.encoding)
	}
	if nil != err {
		t.Fatalf("NewWriterSize failed: %v", err)
	}
	return irWriter
}
