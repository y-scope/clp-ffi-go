#include "serializer.h"

#include <string_view>
#include <vector>

#include <clp/ffi/ir_stream/decoding_methods.hpp>
#include <clp/ffi/ir_stream/encoding_methods.hpp>
#include <clp/ir/types.hpp>
#include <clp/time_types.hpp>

#include "ffi_go/api_decoration.h"
#include "ffi_go/defs.h"
#include "ffi_go/ir/types.hpp"

namespace ffi_go::ir {
using clp::ffi::ir_stream::IRErrorCode;
using clp::ir::eight_byte_encoded_variable_t;
using clp::ir::four_byte_encoded_variable_t;

namespace {
/**
 * Generic helper for ir_serializer_new_*_serializer_with_preamble functions.
 */
template <class encoded_variable_t>
[[nodiscard]] auto new_serializer_with_preamble(
        StringView ts_pattern,
        StringView ts_pattern_syntax,
        StringView time_zone_id,
        [[maybe_unused]] epoch_time_ms_t reference_ts,
        void** ir_serializer_ptr,
        ByteSpan* ir_view
) -> int;

/**
 * Generic helper for ir_serializer_serialize_*_log_event functions.
 */
template <class encoded_variable_t>
[[nodiscard]] auto serialize_log_event(
        StringView log_message,
        epoch_time_ms_t timestamp_or_delta,
        void* ir_serializer,
        ByteSpan* ir_view
) -> int;

template <class encoded_variable_t>
auto new_serializer_with_preamble(
        StringView ts_pattern,
        StringView ts_pattern_syntax,
        StringView time_zone_id,
        [[maybe_unused]] epoch_time_ms_t reference_ts,
        void** ir_serializer_ptr,
        ByteSpan* ir_view
) -> int {
    if (nullptr == ir_serializer_ptr || nullptr == ir_view) {
        return static_cast<int>(IRErrorCode::IRErrorCode_Corrupted_IR);
    }
    Serializer* serializer{new Serializer{}};
    if (nullptr == serializer) {
        return static_cast<int>(IRErrorCode::IRErrorCode_Corrupted_IR);
    }
    *ir_serializer_ptr = serializer;

    bool success{false};
    if constexpr (std::is_same_v<encoded_variable_t, eight_byte_encoded_variable_t>) {
        success = clp::ffi::ir_stream::eight_byte_encoding::serialize_preamble(
                std::string_view{ts_pattern.m_data, ts_pattern.m_size},
                std::string_view{ts_pattern_syntax.m_data, ts_pattern_syntax.m_size},
                std::string_view{time_zone_id.m_data, time_zone_id.m_size},
                serializer->m_ir_buf
        );
    } else if constexpr (std::is_same_v<encoded_variable_t, four_byte_encoded_variable_t>) {
        success = clp::ffi::ir_stream::four_byte_encoding::serialize_preamble(
                std::string_view{ts_pattern.m_data, ts_pattern.m_size},
                std::string_view{ts_pattern_syntax.m_data, ts_pattern_syntax.m_size},
                std::string_view{time_zone_id.m_data, time_zone_id.m_size},
                reference_ts,
                serializer->m_ir_buf
        );
    } else {
        static_assert(cAlwaysFalse<encoded_variable_t>, "Invalid/unhandled encoding type");
    }
    if (false == success) {
        return static_cast<int>(IRErrorCode::IRErrorCode_Corrupted_IR);
    }

    ir_view->m_data = serializer->m_ir_buf.data();
    ir_view->m_size = serializer->m_ir_buf.size();
    return static_cast<int>(IRErrorCode::IRErrorCode_Success);
}

template <class encoded_variable_t>
auto serialize_log_event(
        StringView log_message,
        epoch_time_ms_t timestamp_or_delta,
        void* ir_serializer,
        ByteSpan* ir_view
) -> int {
    if (nullptr == ir_serializer || nullptr == ir_view) {
        return static_cast<int>(IRErrorCode::IRErrorCode_Corrupted_IR);
    }
    Serializer* serializer{static_cast<Serializer*>(ir_serializer)};
    serializer->m_ir_buf.clear();
    serializer->reserve(log_message.m_size);

    bool success{false};
    if constexpr (std::is_same_v<encoded_variable_t, eight_byte_encoded_variable_t>) {
        success = clp::ffi::ir_stream::eight_byte_encoding::serialize_log_event(
                timestamp_or_delta,
                std::string_view{log_message.m_data, log_message.m_size},
                serializer->m_logtype,
                serializer->m_ir_buf
        );
    } else if constexpr (std::is_same_v<encoded_variable_t, four_byte_encoded_variable_t>) {
        success = clp::ffi::ir_stream::four_byte_encoding::serialize_log_event(
                timestamp_or_delta,
                std::string_view{log_message.m_data, log_message.m_size},
                serializer->m_logtype,
                serializer->m_ir_buf
        );
    } else {
        static_assert(cAlwaysFalse<encoded_variable_t>, "Invalid/unhandled encoding type");
    }
    if (false == success) {
        return static_cast<int>(IRErrorCode::IRErrorCode_Corrupted_IR);
    }

    ir_view->m_data = serializer->m_ir_buf.data();
    ir_view->m_size = serializer->m_ir_buf.size();
    return static_cast<int>(IRErrorCode::IRErrorCode_Success);
}
}  // namespace

CLP_FFI_GO_METHOD auto ir_serializer_close(void* ir_serializer) -> void {
    // NOLINTNEXTLINE(cppcoreguidelines-owning-memory)
    delete static_cast<Serializer*>(ir_serializer);
}

CLP_FFI_GO_METHOD auto ir_serializer_new_eight_byte_serializer_with_preamble(
        StringView ts_pattern,
        StringView ts_pattern_syntax,
        StringView time_zone_id,
        void** ir_serializer_ptr,
        ByteSpan* ir_view
) -> int {
    return new_serializer_with_preamble<eight_byte_encoded_variable_t>(
            ts_pattern,
            ts_pattern_syntax,
            time_zone_id,
            0,
            ir_serializer_ptr,
            ir_view
    );
}

CLP_FFI_GO_METHOD auto ir_serializer_new_four_byte_serializer_with_preamble(
        StringView ts_pattern,
        StringView ts_pattern_syntax,
        StringView time_zone_id,
        epoch_time_ms_t reference_ts,
        void** ir_serializer_ptr,
        ByteSpan* ir_view
) -> int {
    return new_serializer_with_preamble<four_byte_encoded_variable_t>(
            ts_pattern,
            ts_pattern_syntax,
            time_zone_id,
            reference_ts,
            ir_serializer_ptr,
            ir_view
    );
}

CLP_FFI_GO_METHOD auto ir_serializer_serialize_eight_byte_log_event(
        StringView log_message,
        epoch_time_ms_t timestamp,
        void* ir_serializer,
        ByteSpan* ir_view
) -> int {
    return serialize_log_event<eight_byte_encoded_variable_t>(
            log_message,
            timestamp,
            ir_serializer,
            ir_view
    );
}

CLP_FFI_GO_METHOD auto ir_serializer_serialize_four_byte_log_event(
        StringView log_message,
        epoch_time_ms_t timestamp_delta,
        void* ir_serializer,
        ByteSpan* ir_view
) -> int {
    return serialize_log_event<four_byte_encoded_variable_t>(
            log_message,
            timestamp_delta,
            ir_serializer,
            ir_view
    );
}

CLP_FFI_GO_METHOD auto ir_serializer_serialize_utc_offset_change(
        epoch_time_ms_t utc_offset_change,
        void* ir_serializer,
        ByteSpan* ir_view
) -> void {
    Serializer* serializer{static_cast<Serializer*>(ir_serializer)};
    clp::UtcOffset const utc_offset{utc_offset_change};
    serializer->m_ir_buf.clear();
    if (utc_offset != serializer->m_curr_utc_offset) {
        clp::ffi::ir_stream::serialize_utc_offset_change(utc_offset, serializer->m_ir_buf);
        serializer->m_curr_utc_offset = utc_offset;
    }
    ir_view->m_data = serializer->m_ir_buf.data();
    ir_view->m_size = serializer->m_ir_buf.size();
}
}  // namespace ffi_go::ir
