#include "wildcard_query.h"

#include <string>
#include <string_utils/string_utils.hpp>
#include <string_view>

#include "ffi_go/api_decoration.h"
#include "ffi_go/defs.h"

namespace ffi_go::search {
CLP_FFI_GO_METHOD auto wildcard_query_new(StringView query, void** ptr) -> StringView {
    auto* clean{new std::string{clp::string_utils::clean_up_wildcard_search_string(
            std::string_view{query.m_data, query.m_size}
    )}};
    *ptr = clean;
    return {clean->data(), clean->size()};
}

CLP_FFI_GO_METHOD auto wildcard_query_delete(void* str) -> void {
    // NOLINTNEXTLINE(cppcoreguidelines-owning-memory)
    delete static_cast<std::string*>(str);
}

CLP_FFI_GO_METHOD auto wildcard_query_match(StringView target, WildcardQueryView query) -> int {
    return static_cast<int>(clp::string_utils::wildcard_match_unsafe(
            {target.m_data, target.m_size},
            {query.m_query.m_data, query.m_query.m_size},
            query.m_case_sensitive
    ));
}
}  // namespace ffi_go::search
