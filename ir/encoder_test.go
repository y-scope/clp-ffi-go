package ir

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/y-scope/clp-ffi-go/ffi"
)

type WriteCloser interface {
	io.Writer
	io.Closer
}

func openIrEncoder(t *testing.T, eightByte bool) (WriteCloser, IrEncoder) {
	f, err := os.Create(fmt.Sprintf("../testdata/%s.clp", t.Name()))
	if err != nil {
		t.Fatalf("os.Create: %v", err)
	}

	timestampPattern := "yyyy-MM-dd HH:mm:ss,SSS"
	timestampPatternSyntax := "java::SimpleDateFormat"
	timeZoneId := "America/Toronto"

	var irEncoder IrEncoder
	var preamble []byte
	var ret int
	if eightByte {
		var ebIrs EightByteIrStream
		ebIrs, preamble, ret = EightByteEncodePreambleUnsafe(timestampPattern,
			timestampPatternSyntax, timeZoneId)
		irEncoder = &ebIrs
	} else {
		var fbIrs FourByteIrStream
		fbIrs, preamble, ret = FourByteEncodePreambleUnsafe(timestampPattern,
			timestampPatternSyntax, timeZoneId, ffi.EpochTimeMs(time.Now().UnixMilli()))
		irEncoder = &fbIrs
	}
	if 0 != ret {
		t.Fatalf("*EncodePreamble failed: %v", ret)
	}
	n, err := f.Write(preamble)
	if n != len(preamble) {
		t.Fatalf("short write for preamble: %v/%v", n, len(preamble))
	}
	if err != nil {
		t.Fatalf("io.Writer.Write preamble: %v", err)
	}
	return f, irEncoder
}

func writeIrEncoder(t *testing.T, writer io.Writer, irs IrEncoder) {
	msg, ret := irs.EncodeMessageUnsafe(ffi.EpochTimeMs(time.Now().UnixMilli()), "log")
	if 0 != ret {
		t.Fatalf("EncodeMessageUnsafe failed: %v", ret)
	}
	n, err := writer.Write(msg)
	if n != len(msg) {
		t.Fatalf("short write for message: %v/%v", n, len(msg))
	}
	if err != nil {
		t.Fatalf("io.Writer.Write message: %v", err)
	}
}

func TestUnsafeFourByteIrEncoder(t *testing.T) {
	writer, irEncoder := openIrEncoder(t, false)
	defer writer.Close()
	writeIrEncoder(t, writer, irEncoder)
}
