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
	messages []string,
) {
	ioWriter := openIoWriter(t, args)
	irWriter := openIrWriter(t, args, ioWriter)

	var events []ffi.LogEvent
	for _, msg := range messages {
		event := ffi.LogEvent{
			"LogMessage": msg,
			"Timestamp":  time.Now().UnixMilli(),
		}
		_, err := irWriter.WriteLogEvent(event)
		if nil != err {
			t.Fatalf("ir.Writer.WriteLogEvent failed: %v", err)
		}
		events = append(events, event)
	}
	err := irWriter.Close()
	if nil != err {
		t.Fatalf("ir.Writer.Close failed: %v", err)
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
		assertIrLogEvent(t, ioReader, irReader, event)
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
		irWriter, err = NewWriter[EightByteEncoding](writer)
	case fourByteEncoding:
		irWriter, err = NewWriter[FourByteEncoding](writer)
	default:
		t.Fatalf("unsupported encoding: %v", args.encoding)
	}
	if nil != err {
		t.Fatalf("NewWriter failed: %v", err)
	}
	return irWriter
}
