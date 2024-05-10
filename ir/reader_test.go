package ir

import (
	"math"
	"os"
	"testing"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/y-scope/clp-ffi-go/ffi"
	"github.com/y-scope/clp-ffi-go/search"
)

func TestIrReader(t *testing.T) {
	var fpath string = os.Getenv("go_test_ir")
	if "" == fpath {
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

	interval := search.TimestampInterval{Lower: 0, Upper: math.MaxInt64}
	queries := []search.WildcardQuery{
		search.NewWildcardQuery("*ERROR*", true),
		search.NewWildcardQuery("*WARN*", true),
	}
	for {
		var log *ffi.LogEventView
		// log, err = irr.Read()
		// log, err = irr.ReadToContains("ERROR")
		// var _ search.WildcardQuery
		log, _, err = irr.ReadToWildcardMatchWithTimeInterval(
			queries,
			interval,
		)
		if nil != err {
			break
		}
		t.Logf("msg: %v | %v", time.UnixMilli(int64(log.Timestamp)), log.LogMessageView)
	}
	if EndOfIr != err {
		t.Fatalf("Reader.Read failed: %v", err)
	}
}
