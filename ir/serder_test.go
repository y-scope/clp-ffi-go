package ir

import (
	"io"
	"testing"
	"time"

	"github.com/y-scope/clp-ffi-go/ffi"
)

func TestPreamble(t *testing.T) {
	for _, args := range generateTestArgs(t, t.Name()) {
		args := args // capture range variable for func literal
		t.Run(args.name, func(t *testing.T) { t.Parallel(); testPreamble(t, args) })
	}
}

func testPreamble(t *testing.T, args testArgs) {
	writer := openIoWriter(t, args)
	irSerializer := serializeIrPreamble(t, args, writer)

	writer.Close()
	irSerializer.Close()

	reader := openIoReader(t, args)
	assertIrPreamble(t, args, reader)
}

func testSerDerLogMessages(
	t *testing.T,
	args testArgs,
	logMessages []string,
) {
	ioWriter := openIoWriter(t, args)
	irSerializer := serializeIrPreamble(t, args, ioWriter)

	var events []ffi.LogEvent
	for _, msg := range logMessages {
		event := ffi.LogEvent{
			"LogMessage": msg,
			"Timestamp":  uint64(time.Now().UnixMilli()),
		}
		irView, err := irSerializer.SerializeLogEvent(event)
		if nil != err {
			t.Fatalf("SerializeLogEvent failed: %v", err)
		}
		_, err = ioWriter.Write(irView)
		if nil != err {
			t.Fatalf("io.Writer.Write message: %v", err)
		}
		events = append(events, event)
	}
	irSerializer.Close()
	_, err := ioWriter.Write([]byte{0x0})
	if nil != err {
		t.Fatalf("io.Writer.Write message: %v", err)
	}
	ioWriter.Close()

	ioReader := openIoReader(t, args)
	defer ioReader.Close()
	irReader := assertIrPreamble(t, args, ioReader)
	defer irReader.Close()

	for _, event := range events {
		assertIrLogEvent(t, ioReader, irReader, event)
	}
	assertEndOfIr(t, ioReader, irReader)
}

func serializeIrPreamble(
	t *testing.T,
	args testArgs,
	writer io.Writer,
) Serializer {
	var err error
	var serializer Serializer
	var preambleIr BufView
	switch args.encoding {
	case eightByteEncoding:
		serializer, preambleIr, err = EightByteSerializer()
	case fourByteEncoding:
		serializer, preambleIr, err = FourByteSerializer()
	default:
		t.Fatalf("unsupported encoding: %v", args.encoding)
	}
	if nil != err {
		t.Fatalf("constructor failed: %v", err)
	}
	n, err := writer.Write(preambleIr)
	if n != len(preambleIr) {
		t.Fatalf("short write for preamble: %v/%v", n, len(preambleIr))
	}
	if nil != err {
		t.Fatalf("io.Writer.Write preamble: %v", err)
	}
	return serializer
}

func assertIrPreamble(
	t *testing.T,
	args testArgs,
	reader io.Reader,
) *Reader {
	irreader, err := NewReaderSize(reader, 4096)
	if nil != err {
		t.Fatalf("NewReader failed: %v", err)
	}
	return irreader
}
