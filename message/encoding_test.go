package message

import (
	"encoding/binary"
	_ "fmt"
	"math"
	"runtime"
	"testing"

	"github.com/y-scope/clp-ffi-go/test"
)

type testLog struct {
	name string
	msg  string
	// vars []byte
	// dictVars []string
	// dictVarEndOffsets []int32
}

var testlogs []testLog = []testLog{
	{name: "static", msg: "static text static text static text"},
	{name: "int", msg: "0 1 2 3 0123"},
	{name: "float", msg: "0.0 1.1 2.2 3.3 01234.0123"},
	{name: "dict", msg: "dictVar0 dictVar1 dictVar=dictVar2"},
	{name: "combo", msg: "Static text, dictVar1, 123, 456.7, dictVar2, 987, 654.3"},
}

func assertDecodedMessage(t *testing.T, log testLog, err error, msg string) {
	t.Helper()
	if nil != err {
		t.Fatalf("DecodeMessage: %v", err)
	}
	if log.msg != msg {
		t.Fatalf("Test msg does not match LogMessage.Msg:\nwant| %v\ngot| %v", log.msg, msg)
	}
}

func assertEncodedMessage(t *testing.T, log testLog, ret int, logtype []byte, vars []byte) {
	t.Helper()
	if 0 != ret {
		t.Fatalf("EncodeMessage: %v", ret)
	}
	// TODO: test other fields...?
}

func testDecodeMessage(t *testing.T, testlog testLog) {
	uem, ret := EncodeMessageUnsafe(testlog.msg)
	var em EncodedMessage = uem.MakeSafe()
	assertEncodedMessage(t, testlog, ret, em.Logtype, em.Vars)

	log, err := em.unsafeRef.DecodeMessage()
	runtime.GC()

	// calling ReleaseRef allows uem to be collected despite em and msg being
	// still reachable
	em.ReleaseRef()
	test.AssertFinalizers(t, test.NewFinalizer(&uem))
	runtime.GC()

	assertDecodedMessage(t, testlog, err, string(log.Msg))
	test.AssertFinalizers(t, test.NewFinalizer(&em), test.NewFinalizer(&log))
}

func testUnsafeDecodeMessage(t *testing.T, testlog testLog) {
	em, ret := EncodeMessageUnsafe(testlog.msg)
	assertEncodedMessage(t, testlog, ret, em.Logtype, em.Vars)

	log, err := em.DecodeMessage()
	runtime.GC()
	assertDecodedMessage(t, testlog, err, string(log.Msg))
	test.AssertFinalizers(t, test.NewFinalizer(&em), test.NewFinalizer(&log))
}

func TestSafeEncodeDecodeMessage(t *testing.T) {
	for _, testlog := range testlogs {
		test := testlog
		t.Run(test.name, func(t *testing.T) { testDecodeMessage(t, test) })
	}
}

func TestUnsafeEncodeDecodeMessage(t *testing.T) {
	for _, testlog := range testlogs {
		test := testlog
		t.Run(test.name, func(t *testing.T) { testUnsafeDecodeMessage(t, test) })
	}
}

func Float64frombytes(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}

func Float64bytes(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}
