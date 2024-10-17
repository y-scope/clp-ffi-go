package ir

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/klauspost/compress/zstd"

	"github.com/y-scope/clp-ffi-go/ffi"
)

type testArg int

const (
	eightByteEncoding testArg = iota
	fourByteEncoding
	noCompression
	zstdCompression
)

var testArgStr = []string{
	"eightByteEncoding",
	"fourByteEncoding",
	"noCompression",
	"zstdCompression",
}

type testArgs struct {
	encoding    testArg
	compression testArg
	name        string
	filePath    string
}

var testKeys = []string{
	"LogMessage",
	"Timestamp",
}

func TestLogMessagesCombo(t *testing.T) {
	messages := []string{
		"static text dict=var notint123 -1.234 4321.",
		"static123 text321 dict=var0123 321.1234 -3210.",
	}
	testLogMessages(t, messages)
}

func TestLogMessagesDict(t *testing.T) {
	messages := []string{
		"textint1234 textequal=variable",
		fmt.Sprintf("test=bigint %v", math.MaxInt32+1),
	}
	testLogMessages(t, messages)
}

func TestLogMessagesFloat(t *testing.T) {
	messages := []string{
		"float 1.0 1.2 1.23 1.234",
		"-float -1.0 -1.2 -1.23 -1.234",
	}
	testLogMessages(t, messages)
}

func TestLogMessagesInt(t *testing.T) {
	messages := []string{
		"int 1 12 123 1234",
		"-int -1 -12 -123 -1234",
	}
	testLogMessages(t, messages)
}

func TestLogMessagesStatic(t *testing.T) {
	messages := []string{
		"static text log zero.",
		"static text log one.",
	}
	testLogMessages(t, messages)
}

func TestLogMessagesLongLogs(t *testing.T) {
	const eightMB int = 8 * 1024 * 1024
	messages := []string{
		strings.Repeat("x", eightMB),
		strings.Repeat("x", eightMB-1),
	}
	testLogMessages(t, messages)
}

func assertEndOfIr(
	t *testing.T,
	reader io.Reader,
	irReader *Reader,
) {
	_, err := irReader.Read()
	if EndOfIr != err {
		t.Fatalf("assertEndOfIr failed got: %v", err)
	}
}

func assertIrLogEvent(
	t *testing.T,
	reader io.Reader,
	irReader *Reader,
	refEvent ffi.LogEvent,
) {
	event, err := irReader.Read()
	if nil != err {
		t.Fatalf("Reader.Read failed: %v", err)
	}
	for _, key := range testKeys {
		if refEvent[key] != event[key] {
			t.Fatalf("Reader.Read wrong %v, wanted: '%v' got: '%v'",
				key,
				refEvent[key],
				event[key],
			)
		}
	}
	t.Logf("'%v'\n", event)
}

func generateTestArgs(t *testing.T, prefix string) []testArgs {
	var tests []testArgs
	tmpdir := t.TempDir()
	for _, encoding := range []testArg{eightByteEncoding, fourByteEncoding} {
		for _, compression := range []testArg{noCompression, zstdCompression} {
			testName := prefix + "-" + testArgStr[encoding] + "-" + testArgStr[compression]
			fileName := testName + ".clp"
			if zstdCompression == compression {
				fileName += ".zst"
			}
			filePath := filepath.Join(tmpdir, fileName)
			tests = append(tests, testArgs{encoding, compression, testName, filePath})
		}
	}
	return tests
}

func testLogMessages(t *testing.T, messages []string) {
	for _, args := range generateTestArgs(t, t.Name()+"-SerDer") {
		args := args // capture range variable for func literal
		t.Run(
			args.name,
			func(t *testing.T) { t.Parallel(); testSerDerLogMessages(t, args, messages) },
		)
	}
	for _, args := range generateTestArgs(t, t.Name()+"-WriteRead") {
		args := args // capture range variable for func literal
		t.Run(
			args.name,
			func(t *testing.T) { t.Parallel(); testWriteReadLogMessages(t, args, messages) },
		)
	}
}

func openIoReader(t *testing.T, args testArgs) io.ReadCloser {
	file, err := os.Open(args.filePath)
	if nil != err {
		t.Fatalf("os.Open: %v", err)
	}
	var reader io.ReadCloser
	switch args.compression {
	case noCompression:
		reader = file
	case zstdCompression:
		reader, err = newZstdReader(file)
		if nil != err {
			t.Fatalf("zstd.NewReader failed: %v", err)
		}
	default:
		t.Fatalf("unsupported compression: %v", args.compression)
	}
	return reader
}

func openIoWriter(t *testing.T, args testArgs) io.WriteCloser {
	file, err := os.Create(args.filePath)
	if nil != err {
		t.Fatalf("os.Create: %v", err)
	}
	var writer io.WriteCloser
	switch args.compression {
	case noCompression:
		writer = file
	case zstdCompression:
		writer, err = zstd.NewWriter(file)
		if nil != err {
			t.Fatalf("zstd.NewWriter failed: %v", err)
		}
	default:
		t.Fatalf("unsupported compression: %v", args.compression)
	}
	return writer
}

type zstdReader struct {
	*zstd.Decoder
}

func newZstdReader(reader io.Reader) (*zstdReader, error) {
	zreader, err := zstd.NewReader(reader)
	return &zstdReader{zreader}, err
}

func (reader *zstdReader) Close() error {
	reader.Decoder.Close()
	return nil
}
