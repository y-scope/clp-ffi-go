package search

/*
#include <ffi_go/defs.h>
#include <ffi_go/search/wildcard_query.h>
*/
import "C"

import (
	"strings"
	"unsafe"

	"github.com/y-scope/clp-ffi-go/ffi"
)

// A CLP wildcard query containing a query string and a bool for whether the
// query is case sensitive or not. The fields must be accessed through getters
// to ensure that the query string remains clean/safe after creation by
// NewWildcardQuery.
// (Copied from clp/components/core/src/string_utils.hpp)
// Two wildcards are currently supported: '*' to match 0 or more characters, and
// '?' to match any single character. Each can be escaped using a preceding '\'.
// Other characters which are escaped are treated as normal characters.
type WildcardQuery struct {
	query         string
	caseSensitive bool
}

// Create a new WildcardQuery that is cleaned to contain a safe wildcard query
// string. A wildcard query string must follow 2 rules:
//  1. The wildcard string should not contain consecutive '*'.
//  2. The wildcard string should not contain an escape character without a
//     character following it.
//
// NewWildcardQuery will sanitize the provided query and store the safe version.
func NewWildcardQuery(query string, caseSensitive bool) WildcardQuery {
	var cptr unsafe.Pointer
	cleanQuery := C.wildcard_query_clean(
		C.StringView{
			(*C.char)(unsafe.Pointer(unsafe.StringData(query))),
			C.size_t(len(query)),
		},
		&cptr,
	)
	defer C.wildcard_query_delete(cptr)
	return WildcardQuery{
		strings.Clone(unsafe.String(
			(*byte)((unsafe.Pointer)(cleanQuery.m_data)),
			cleanQuery.m_size,
		)),
		caseSensitive,
	}
}

func (self WildcardQuery) Query() string       { return self.query }
func (self WildcardQuery) CaseSensitive() bool { return self.caseSensitive }

// A MergedWildcardQuery represents the union of multiple wildcard queries
// (multiple WildcardQuery instances each with their own query string and case
// sensitivity).
type MergedWildcardQuery struct {
	queries         string
	endOffsets      []int
	caseSensitivity []bool
}

func (self MergedWildcardQuery) Queries() string         { return self.queries }
func (self MergedWildcardQuery) EndOffsets() []int       { return self.endOffsets }
func (self MergedWildcardQuery) CaseSensitivity() []bool { return self.caseSensitivity }

// Merge multiple WildcardQuery objects together by concatenating their query
// strings, storing their end/length offsets, and recording their case
// sensitivity.
func MergeWildcardQueries(queries []WildcardQuery) MergedWildcardQuery {
	var sb strings.Builder
	offsets := make([]int, len(queries))
	caseSensitivity := make([]bool, len(queries))
	for i, q := range queries {
		offsets[i], _ = sb.WriteString(q.query) // err always nil
		caseSensitivity[i] = queries[i].caseSensitive
	}
	return MergedWildcardQuery{sb.String(), offsets, caseSensitivity}
}

// A timestamp interval of [m_lower, m_upper).
type TimestampInterval struct {
	Lower ffi.EpochTimeMs
	Upper ffi.EpochTimeMs
}
