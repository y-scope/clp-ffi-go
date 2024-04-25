#include "wildcard_query.h"

#include <string>
#include <string_view>

#include <clp/components/core/src/string_utils.hpp>

#include <ffi_go/defs.h>

namespace ffi_go::search {
extern "C" auto wildcard_query_delete(void* str) -> void {
    // NOLINTNEXTLINE(cppcoreguidelines-owning-memory)
    delete static_cast<std::string*>(str);
}

extern "C" auto wildcard_query_clean(StringView query, void** ptr) -> StringView {
    auto* clean{new std::string{
            clean_up_wildcard_search_string(std::string_view{query.m_data, query.m_size})
    }};
    *ptr = clean;
    return {clean->data(), clean->size()};
}

extern "C" auto wildcard_query_match(StringView target, WildcardQueryView query) -> int {
    return static_cast<int>(wildcard_match_unsafe(
            std::string_view{target.m_data, target.m_size},
            std::string_view{query.m_query.m_data, query.m_query.m_size},
            static_cast<bool>(query.m_case_sensitive)
    ));
}
}  // namespace ffi_go::search
