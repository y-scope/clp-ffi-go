package ir

import (
	"os"
	"strings"
	"testing"

	"github.com/klauspost/compress/zstd"

	"github.com/y-scope/clp-ffi-go/ffi"
)

func TestIrReaderOnFile(t *testing.T) {
	var fpath string = os.Getenv("go_test_ir")
	if fpath == "" {
		t.Skip("Set an input ir stream using the env variable: go_test_ir")
	}
	var err error
	var file *os.File
	if file, err = os.Open(fpath); nil != err {
		t.Fatalf("os.Open failed: %v", err)
	}
	defer file.Close()

	reader, _ := zstd.NewReader(file)
	defer reader.Close()

	var irr *Reader
	if irr, err = NewReaderSize(reader, 512*1024*1024); nil != err {
		t.Fatalf("NewReader failed: %v", err)
	}
	defer irr.Close()

	for {
		var event ffi.LogEvent
		// event, err = irr.Read()
		event, err = irr.ReadToFunc(func(event ffi.LogEvent) bool {
			return strings.Contains(event["message"].(string), "ERROR")
		})
		if nil != err {
			break
		}
		t.Logf("msg: %v", event["message"])
	}
	if EndOfIr != err {
		t.Fatalf("Reader.Read failed: %v", err)
	}
}
