#ifndef FFI_GO_IR_WILDCARD_QUERY_H
#define FFI_GO_IR_WILDCARD_QUERY_H
// header must support C, making modernize checks inapplicable
// NOLINTBEGIN(modernize-deprecated-headers)
// NOLINTBEGIN(modernize-use-trailing-return-type)
// NOLINTBEGIN(modernize-use-using)

#include <stdbool.h>

#include <ffi_go/defs.h>

#ifdef __cplusplus
extern "C" {
#endif

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
     * Delete a std::string holding a wildcard query.
     * @param[in] str Address of a std::string created and returned by
     *   clean_wildcard_query
     */
    void wildcard_query_delete(void* str);

    /**
     * Given a query string, clean it to be safe for matching. See
     * `clean_up_wildcard_search_string` in CLP src/string_utils.hpp.
     * @param[in] query Query string to clean
     * @param[in] ptr Address of a new std::string
     * @return New string holding cleaned query
     */
    StringView wildcard_query_clean(StringView query, void** ptr);

    /**
     * Given a target string perform CLP wildcard matching using query. See
     * `wildcard_match_unsafe` in CLP src/string_utils.hpp.
     * @param[in] target String to perform matching on
     * @param[in] query Query to use for matching
     * @return 1 if query matches target, 0 otherwise
     */
    int wildcard_query_match(StringView target, WildcardQueryView query);

#ifdef __cplusplus
}
#endif

// NOLINTEND(modernize-use-using)
// NOLINTEND(modernize-use-trailing-return-type)
// NOLINTEND(modernize-deprecated-headers)
#endif  // FFI_GO_IR_WILDCARD_QUERY_H
