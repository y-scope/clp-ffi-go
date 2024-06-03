#ifndef FFI_GO_IR_WILDCARD_QUERY_H
#define FFI_GO_IR_WILDCARD_QUERY_H
// header must support C, making modernize checks inapplicable
// NOLINTBEGIN(modernize-deprecated-headers)
// NOLINTBEGIN(modernize-use-trailing-return-type)
// NOLINTBEGIN(modernize-use-using)

#include <stdbool.h>

#include "ffi_go/api_decoration.h"
#include "ffi_go/defs.h"

/**
 * A timestamp interval of [m_lower, m_upper).
 */
typedef struct {
    epoch_time_ms_t m_lower;
    epoch_time_ms_t m_upper;
} TimestampInterval;

/**
 * A view of a wildcard query passed down from Go. The query string is assumed
 * to have been cleaned using the CLP function `clean_up_wildcard_search_string`
 * on construction. m_case_sensitive is 1 for a case sensitive query (0 for case
 * insensitive).
 */
typedef struct {
    StringView m_query;
    bool m_case_sensitive;
} WildcardQueryView;

/**
 * A view of a Go search.MergedWildcardQuery passed down through Cgo. The
 * string is a concatenation of all wildcard queries, while m_end_offsets stores
 * the size of each query.
 */
typedef struct {
    StringView m_queries;
    SizetSpan m_end_offsets;
    BoolSpan m_case_sensitivity;
} MergedWildcardQueryView;

/**
 * Given a query string, allocate and return a clean string that is safe for
 * matching. See `clean_up_wildcard_search_string` in CLP for more details.
 * @param[in] query Query string to clean
 * @param[in] ptr Address of a new std::string
 * @return New string holding cleaned query
 */
CLP_FFI_GO_METHOD StringView wildcard_query_new(StringView query, void** ptr);

/**
 * Delete a std::string holding a wildcard query.
 * @param[in] str Address of a std::string created and returned by
 *   clean_wildcard_query
 */
CLP_FFI_GO_METHOD void wildcard_query_delete(void* str);

/**
 * Given a target string perform CLP wildcard matching using query. See
 * `wildcard_match_unsafe` in CLP src/string_utils.hpp.
 * @param[in] target String to perform matching on
 * @param[in] query Query to use for matching
 * @return 1 if query matches target, 0 otherwise
 */
CLP_FFI_GO_METHOD int wildcard_query_match(StringView target, WildcardQueryView query);

// NOLINTEND(modernize-use-using)
// NOLINTEND(modernize-use-trailing-return-type)
// NOLINTEND(modernize-deprecated-headers)
#endif  // FFI_GO_IR_WILDCARD_QUERY_H
