#include "decoder.h"

#include <string>
#include <string_view>
#include <type_traits>

#include <clp/components/core/src/clp/ffi/encoding_methods.hpp>
#include <clp/components/core/src/clp/ffi/ir_stream/decoding_methods.hpp>
#include <clp/components/core/src/clp/ir/types.hpp>

#include "ffi_go/api_decoration.h"
#include "ffi_go/defs.h"
#include "ffi_go/ir/types.hpp"

namespace ffi_go::ir {
using clp::ffi::ir_stream::IRErrorCode;

namespace {
/**
 * Generic helper for ir_decoder_decode_*_log_message
 */
template <class encoded_var_view_t>
[[nodiscard]] auto decode_log_message(
        StringView logtype,
        encoded_var_view_t vars,
        StringView dict_vars,
        Int32tSpan dict_var_end_offsets,
        void* ir_decoder,
        StringView* log_msg_view
) -> int {
    using encoded_var_t = std::conditional_t<
            std::is_same_v<Int64tSpan, encoded_var_view_t>,
            clp::ir::eight_byte_encoded_variable_t,
            clp::ir::four_byte_encoded_variable_t>;
    if (nullptr == ir_decoder || nullptr == log_msg_view) {
        return static_cast<int>(IRErrorCode::IRErrorCode_Corrupted_IR);
    }
    Decoder* decoder{static_cast<Decoder*>(ir_decoder)};
    auto& log_msg{decoder->m_log_message};
    log_msg.reserve(logtype.m_size + dict_vars.m_size);

    IRErrorCode err{IRErrorCode::IRErrorCode_Success};
    try {
        log_msg = clp::ffi::decode_message<encoded_var_t>(
                std::string_view(logtype.m_data, logtype.m_size),
                vars.m_data,
                vars.m_size,
                std::string_view(dict_vars.m_data, dict_vars.m_size),
                dict_var_end_offsets.m_data,
                dict_var_end_offsets.m_size
        );
    } catch (clp::ffi::EncodingException const& e) {
        err = IRErrorCode::IRErrorCode_Decode_Error;
    }

    log_msg_view->m_data = log_msg.data();
    log_msg_view->m_size = log_msg.size();
    return static_cast<int>(err);
}
}  // namespace

CLP_FFI_GO_METHOD auto ir_decoder_new() -> void* {
    // NOLINTNEXTLINE(cppcoreguidelines-owning-memory)
    return new Decoder{};
}

CLP_FFI_GO_METHOD auto ir_decoder_close(void* ir_decoder) -> void {
    // NOLINTNEXTLINE(cppcoreguidelines-owning-memory)
    delete static_cast<Decoder*>(ir_decoder);
}

CLP_FFI_GO_METHOD auto ir_decoder_decode_eight_byte_log_message(
        StringView logtype,
        Int64tSpan vars,
        StringView dict_vars,
        Int32tSpan dict_var_end_offsets,
        void* ir_decoder,
        StringView* log_message
) -> int {
    return decode_log_message(
            logtype,
            vars,
            dict_vars,
            dict_var_end_offsets,
            ir_decoder,
            log_message
    );
}

CLP_FFI_GO_METHOD auto ir_decoder_decode_four_byte_log_message(
        StringView logtype,
        Int32tSpan vars,
        StringView dict_vars,
        Int32tSpan dict_var_end_offsets,
        void* ir_decoder,
        StringView* log_message
) -> int {
    return decode_log_message(
            logtype,
            vars,
            dict_vars,
            dict_var_end_offsets,
            ir_decoder,
            log_message
    );
}
}  // namespace ffi_go::ir
