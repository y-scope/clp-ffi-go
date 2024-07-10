package ir

import (
	"io"
	"testing"
	"time"

	"github.com/y-scope/clp-ffi-go/ffi"
)

func TestPreamble(t *testing.T) {
	preamble := preambleFields{
		TimestampInfo{defaultTimestampPattern, defaultTimestampPatternSyntax, defaultTimeZoneId},
		ffi.EpochTimeMs(time.Now().UnixMilli()),
	}
	for _, args := range generateTestArgs(t, t.Name()) {
		args := args // capture range variable for func literal
		t.Run(args.name, func(t *testing.T) { t.Parallel(); testPreamble(t, args, preamble) })
	}
}

func testPreamble(t *testing.T, args testArgs, preamble preambleFields) {
	writer := openIoWriter(t, args)
	irSerializer := serializeIrPreamble(t, args, preamble, writer)

	writer.Close()
	irSerializer.Close()

	reader := openIoReader(t, args)
	assertIrPreamble(t, args, reader, preamble)
}

func testSerDerLogMessages(
	t *testing.T,
	args testArgs,
	logMessages []ffi.LogMessage,
) {
	ioWriter := openIoWriter(t, args)

	preamble := preambleFields{
		TimestampInfo{defaultTimestampPattern, defaultTimestampPatternSyntax, defaultTimeZoneId},
		ffi.EpochTimeMs(time.Now().UnixMilli()),
	}
	irSerializer := serializeIrPreamble(t, args, preamble, ioWriter)

	utcOffsetToronto := ffi.EpochTimeMs(-4 * 60 * 60 * 1000)
	irView := irSerializer.SerializeUtcOffsetChange(utcOffsetToronto)
	_, err := ioWriter.Write(irView)
	if nil != err {
		t.Fatalf("io.Writer.Write message: %v", err)
	}

	var events []ffi.LogEvent
	for _, msg := range logMessages {
		event := ffi.LogEvent{
			LogMessage: msg,
			Timestamp:  ffi.EpochTimeMs(time.Now().UnixMilli()),
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
	_, err = ioWriter.Write([]byte{0x0})
	if nil != err {
		t.Fatalf("io.Writer.Write message: %v", err)
	}
	ioWriter.Close()

	ioReader := openIoReader(t, args)
	defer ioReader.Close()
	irReader := assertIrPreamble(t, args, ioReader, preamble)
	defer irReader.Close()

	for _, event := range events {
		assertIrLogEvent(t, ioReader, irReader, event, utcOffsetToronto)
	}
	assertEndOfIr(t, ioReader, irReader)
}

func serializeIrPreamble(
	t *testing.T,
	args testArgs,
	preamble preambleFields,
	writer io.Writer,
) Serializer {
	var err error
	var serializer Serializer
	var preambleIr BufView
	switch args.encoding {
	case eightByteEncoding:
		serializer, preambleIr, err = EightByteSerializer(
			preamble.Pattern,
			preamble.PatternSyntax,
			preamble.TimeZoneId,
		)
	case fourByteEncoding:
		serializer, preambleIr, err = FourByteSerializer(
			preamble.Pattern,
			preamble.PatternSyntax,
			preamble.TimeZoneId,
			preamble.prevTimestamp,
		)
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
	preamble preambleFields,
) *Reader {
	irreader, err := NewReaderSize(reader, 4096)
	if nil != err {
		t.Fatalf("NewReader failed: %v", err)
	}
	if irreader.TimestampInfo().Pattern != preamble.Pattern {
		t.Fatalf(
			"NewReader wrong pattern: '%v' != '%v'",
			irreader.TimestampInfo().Pattern,
			preamble.Pattern,
		)
	}
	if irreader.TimestampInfo().PatternSyntax != preamble.PatternSyntax {
		t.Fatalf(
			"NewReader wrong pattern syntax: '%v' != '%v'",
			irreader.TimestampInfo().PatternSyntax,
			preamble.PatternSyntax,
		)
	}
	if irreader.TimestampInfo().TimeZoneId != preamble.TimeZoneId {
		t.Fatalf(
			"NewReader wrong time zone id: '%v' != '%v'",
			irreader.TimestampInfo().TimeZoneId,
			preamble.TimeZoneId,
		)
	}
	if fourByteEncoding == args.encoding {
		deserializer, ok := irreader.Deserializer.(*fourByteDeserializer)
		if false == ok {
			t.Fatalf("casting Deserializer to *fourByteDeserializer failed for fourByteEncoding.")
		}
		if deserializer.prevTimestamp != preamble.prevTimestamp {
			t.Fatalf(
				"NewReader wrong reference timestamp: '%v' != '%v'",
				deserializer.prevTimestamp,
				preamble.prevTimestamp,
			)
		}
	}
	return irreader
}
