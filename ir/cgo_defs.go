package ir

/*
#include <ffi_go/defs.h>
*/
import "C"

import (
	"unsafe"
)

// The follow functions are helpers to cleanup Cgo related code. The underlying
// Go type created from a 'C' type is not exported and recreated in each
// package. Therefore, these helpers must be redefined in any package wishing to
// use them, so that they reference the correct underlying Go type of the
// package (see: https://pkg.go.dev/cmd/cgo). This problem could be alivated by
// using Go generate to create/add these helpers to a package necessary.

func newCByteSpan(s []byte) C.ByteSpan {
	return C.ByteSpan{
		unsafe.Pointer(unsafe.SliceData(s)),
		C.size_t(len(s)),
	}
}

func newCInt32tSpan(s []int32) C.Int32tSpan {
	return C.Int32tSpan{
		(*C.int32_t)(unsafe.Pointer(unsafe.SliceData(s))),
		C.size_t(len(s)),
	}
}

func newCInt64tSpan(s []int64) C.Int64tSpan {
	return C.Int64tSpan{
		(*C.int64_t)(unsafe.Pointer(unsafe.SliceData(s))),
		C.size_t(len(s)),
	}
}

func newCStringView(s string) C.StringView {
	return C.StringView{
		(*C.char)(unsafe.Pointer(unsafe.StringData(s))),
		C.size_t(len(s)),
	}
}

func newLogMessageView[Tgo EightByteEncoding | FourByteEncoding, Tc C.Int64tSpan | C.Int32tSpan](
	logtype C.StringView,
	vars Tc,
	dictVars C.StringView,
	dictVarEndOffsets C.Int32tSpan,
) *LogMessageView[Tgo] {
	var msgView LogMessageView[Tgo]
	msgView.Logtype = unsafe.String((*byte)(unsafe.Pointer(logtype.m_data)), logtype.m_size)
	switch any(msgView.Vars).(type) {
	case []EightByteEncoding:
		dst := any(&msgView.Vars).(*[]EightByteEncoding)
		src := any(vars).(C.Int64tSpan)
		if 0 < src.m_size && nil != src.m_data {
			*dst = unsafe.Slice((*EightByteEncoding)(src.m_data), src.m_size)
		}
	case []FourByteEncoding:
		dst := any(&msgView.Vars).(*[]FourByteEncoding)
		src := any(vars).(C.Int32tSpan)
		if 0 < src.m_size && nil != src.m_data {
			*dst = unsafe.Slice((*FourByteEncoding)(src.m_data), src.m_size)
		}
	default:
		return nil
	}
	if 0 < dictVars.m_size && nil != dictVars.m_data {
		msgView.Logtype = unsafe.String((*byte)(unsafe.Pointer(dictVars.m_data)), dictVars.m_size)
	}
	if 0 < dictVarEndOffsets.m_size && nil != dictVarEndOffsets.m_data {
		msgView.DictVarEndOffsets = unsafe.Slice(
			(*int32)(dictVarEndOffsets.m_data),
			dictVarEndOffsets.m_size,
		)
	}
	return &msgView
}
