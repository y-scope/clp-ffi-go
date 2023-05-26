package message

/*
#include <log_event.h>
#include <message/encoding.h>
*/
import "C"

import (
	"runtime"
	"unsafe"

	"github.com/y-scope/clp-ffi-go/ffi"
)

/* TODO outdated
There are two sets of structs exposed:
1. DecodedMessage, EncodedMessage
The fields of these structs point to regular go memory. This memory was
populated by copying the data returned by the native calls. The behaviour of
these structs is the same as any normal go struct, at the expense of extra
copying.
To greatly simplify decoding an EncodedMessage uses a private reference to the
DecodedMessageUnsafe that created it. Holding this reference prevents the c
memory from being freed and will increase memory usage. ReleaseRef can be used
to drop this reference, but DecodeMessage will always return an error
afterwards.

2. DecodedMessageUnsafe, EncodedMessageUnsafe
The fields of these structs point to c memory. Slices are created to wrap the
native memory, with no copying performed. The underlying c memory will be freed
by a finalizer set on the creation of these objects. This means once the
original object becomes unreachable any access to the underlying memory is
undefined. Any reference created to this memory (e.g. making a copy of the
object or fields, new slices of the fields, etc) is only valid as long as the
original object is reachable.
With that said, if all usage is made through the original object, practical
usage will behave as expected.

There is no common interface to abstract/generalize the use of these structs.
It would be fairly easy for an unsuspecting user to pass an unsafe structure to
a function that will not handle it properly. We encourage the user to be
explicit about the usage of unsafe structs.

*/

type EncodedMessage struct {
	Logtype           []byte
	Vars              []byte
	DictVars          []byte
	DictVarEndOffsets []int32
	unsafeRef         *EncodedMessageUnsafe
}

type EncodedMessageUnsafe struct {
	Logtype           []byte
	Vars              []byte
	DictVars          []byte
	DictVarEndOffsets []int32
	cPtr              unsafe.Pointer
}

func (self *EncodedMessage) ReleaseRef() {
	self.unsafeRef = nil
}

func (self *EncodedMessageUnsafe) MakeSafe() EncodedMessage {
	var em EncodedMessage
	em.unsafeRef = self
	em.Logtype = make([]byte, len(self.Logtype))
	em.Vars = make([]byte, len(self.Vars))
	em.DictVars = make([]byte, len(self.DictVars))
	em.DictVarEndOffsets = make([]int32, len(self.DictVarEndOffsets))
	copy(em.Logtype, self.Logtype)
	copy(em.Vars, self.Vars)
	copy(em.DictVars, self.DictVars)
	copy(em.DictVarEndOffsets, self.DictVarEndOffsets)
	return em
}

func (self *EncodedMessage) DecodeMessage() (ffi.LogMessage, error) {
	if nil == self.unsafeRef {
		return ffi.LogMessage{}, NilRef
	}
	msg, ret := self.unsafeRef.DecodeMessage()
	return msg, ret
}

func (self *EncodedMessageUnsafe) DecodeMessage() (ffi.LogMessage, error) {
	var msgClass unsafe.Pointer
	var msg *C.char
	var msgSize C.size_t
	C.decode_message(self.cPtr, &msgClass, &msg, &msgSize)

	if nil == msgClass || nil == msg {
		return ffi.LogMessage{}, DecodeError
	}

	logmsg := ffi.NewLogMessage(unsafe.Pointer(msg), uint64(msgSize), msgClass)
	return logmsg, nil
}

func EncodeMessage(msg string) (EncodedMessage, int) {
	em, ret := EncodeMessageUnsafe(msg)
	return em.MakeSafe(), ret
}

func EncodeMessageUnsafe(msg string) (EncodedMessageUnsafe, int) {
	var logtypePtr, varsPtr, dictVarsPtr, dictVarEndOffsetsPtr unsafe.Pointer
	var logtypeSize, varsSize, dictVarsSize, dictVarEndOffsetsSize uint64
	var em EncodedMessageUnsafe
	em.cPtr = C.encode_message(unsafe.Pointer(&[]byte(msg)[0]), C.size_t(len(msg)),
		&logtypePtr, unsafe.Pointer(&logtypeSize),
		&varsPtr, unsafe.Pointer(&varsSize),
		&dictVarsPtr, unsafe.Pointer(&dictVarsSize),
		&dictVarEndOffsetsPtr, unsafe.Pointer(&dictVarEndOffsetsSize))
	if nil == em.cPtr {
		return em, -1
	}
	em.Logtype = unsafe.Slice((*byte)(logtypePtr), logtypeSize)
	if nil == em.Logtype {
		return em, -2
	}
	if 0 != varsSize {
		em.Vars = unsafe.Slice((*byte)(varsPtr), varsSize)
		if nil == em.Vars {
			return em, -3
		}
	}
	if 0 != dictVarsSize {
		em.DictVars = unsafe.Slice((*byte)(dictVarsPtr), dictVarsSize)
		if nil == em.DictVars {
			return em, -4
		}
	}
	if 0 != dictVarEndOffsetsSize {
		em.DictVarEndOffsets = unsafe.Slice((*int32)(dictVarEndOffsetsPtr), dictVarEndOffsetsSize)
		if nil == em.DictVarEndOffsets {
			return em, -5
		}
	}
	runtime.SetFinalizer(&em,
		func(em *EncodedMessageUnsafe) { C.delete_encoded_message(em.cPtr) })
	return em, 0
}
