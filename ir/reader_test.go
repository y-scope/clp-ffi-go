//go:build test

package ir

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"
	"runtime"

	"github.com/klauspost/compress/zstd"
	"github.com/y-scope/clp-ffi-go/test"
)

func TestFourByteIrReader(t *testing.T) {
	if 0 == len(os.Args) {
		t.Fatalf("This test requires an input ir stream from -args: %v", os.Args)
	}
	var err error
	var file *os.File
	if file, err = os.Open(os.Args[len(os.Args)-1]); nil != err {
		t.Fatalf("os.Open failed: %v", err)
	}
	defer file.Close()

	reader, _ := zstd.NewReader(file)
	defer reader.Close()

	var irr IrReader
	if irr, err = ReadPreamble(reader, 4096); nil != err {
		t.Fatalf("ReadPreamble failed: %v", err)
	}

	fins := []test.Finalizer{}
	for {
		// log, err := irr.ReadNextLogEvent(reader)
		log, err := irr.ReadToContains(reader, []byte("ERROR"))
		// run GC to try and test that log.Msg isn't freed by finalizer
		runtime.GC()
		if nil == err {
			fmt.Printf("msg: %v | %v", time.UnixMilli(int64(log.Timestamp)), string(log.Msg))
		} else if Eof == err || io.EOF == err {
			break
		} else {
			t.Fatalf("ReadNextLogEvent failed: %v", err)
		}
		fins = append(fins, test.NewFinalizer(&log))
	}
	test.AssertFinalizers(t, fins...)
}
