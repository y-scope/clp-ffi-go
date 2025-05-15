#include "serializer.h"

#include <memory>

#include <clp_ffi_go/api_decoration.h>
#include <clp_ffi_go/defs.h>
#include <clp_ffi_go/ir/ErrorCode.hpp>

#include <clp/ffi/ir_stream/Serializer.hpp>
#include <clp/ir/types.hpp>
#include <msgpack.hpp>
#include <outcome/outcome.hpp>

namespace clp_ffi_go::ir {
using clp::ir::eight_byte_encoded_variable_t;
using clp::ir::four_byte_encoded_variable_t;

namespace {
/**
 * Generic helper for ir_serializer_*_close functions.
 */
template <class encoded_variable_t>
auto serializer_close(void* ir_serializer) -> void;

/**
 * Generic helper for ir_serializer_new_*_serializer_with_preamble functions.
 */
template <class encoded_variable_t>
[[nodiscard]] auto serializer_create(void*& ir_serializer_ptr, ByteSpan* ir_span) -> int;

/**
 * Generic helper for ir_serializer_serialize_*_log_event functions.
 */
template <class encoded_variable_t>
[[nodiscard]] auto serialize_log_event(
        void* ir_serializer,
        ByteSpan auto_kv_msgpack_bytes,
        ByteSpan user_kv_msgpack_bytes,
        ByteSpan* ir_span
) -> int;

template <class encoded_variable_t>
auto serializer_close(void* ir_serializer) -> void {
    std::unique_ptr<clp::ffi::ir_stream::Serializer<encoded_variable_t>>(
            static_cast<clp::ffi::ir_stream::Serializer<encoded_variable_t>*>(ir_serializer)
    );
}

template <class encoded_variable_t>
auto serializer_create(void*& ir_serializer_ptr, ByteSpan* ir_span) -> int {
    if (nullptr != ir_serializer_ptr || nullptr == ir_span) {
        return static_cast<int>(ErrorCode::GoIrError::InvalidArg);
    }
    auto result{clp::ffi::ir_stream::Serializer<encoded_variable_t>::create()};
    if (result.has_failure()) {
        return static_cast<int>(ErrorCode::convert_to_go_ir_error(result.error()));
    }
    auto ir_buf_view{result.value().get_ir_buf_view()};
    ir_span->m_data = ir_buf_view.data();
    ir_span->m_size = ir_buf_view.size();
    auto s = std::make_unique<clp::ffi::ir_stream::Serializer<encoded_variable_t>>(
            std::move(result.value())
    );
    ir_serializer_ptr = s.release();
    return static_cast<int>(ErrorCode::GoIrError::Success);
}

template <class encoded_variable_t>
[[nodiscard]] auto serialize_log_event(
        void* ir_serializer,
        ByteSpan auto_kv_msgpack_bytes,
        ByteSpan user_kv_msgpack_bytes,
        ByteSpan* ir_span
) -> int {
    if (nullptr == ir_serializer || nullptr == ir_span) {
        return static_cast<int>(ErrorCode::GoIrError::InvalidArg);
    }
    auto* serializer{
            static_cast<clp::ffi::ir_stream::Serializer<encoded_variable_t>*>(ir_serializer)
    };
    serializer->clear_ir_buf();

    auto const auto_kv_map{msgpack::unpack(
            static_cast<char const*>(auto_kv_msgpack_bytes.m_data),
            auto_kv_msgpack_bytes.m_size
    )};
    auto const user_kv_map{msgpack::unpack(
            static_cast<char const*>(user_kv_msgpack_bytes.m_data),
            user_kv_msgpack_bytes.m_size
    )};
    // NOLINTNEXTLINE(cppcoreguidelines-pro-type-union-access)
    if (false == serializer->serialize_msgpack_map(auto_kv_map->via.map, user_kv_map->via.map)) {
        return static_cast<int>(ErrorCode::GoIrError::IrSerializeError);
    }

    auto ir_buf_view{serializer->get_ir_buf_view()};
    ir_span->m_data = ir_buf_view.data();
    ir_span->m_size = ir_buf_view.size();
    return static_cast<int>(ErrorCode::GoIrError::Success);
}
}  // namespace

CLP_FFI_GO_METHOD auto ir_serializer_eight_byte_close(void* ir_serializer) -> void {
    serializer_close<eight_byte_encoded_variable_t>(ir_serializer);
}

CLP_FFI_GO_METHOD auto ir_serializer_four_byte_close(void* ir_serializer) -> void {
    serializer_close<four_byte_encoded_variable_t>(ir_serializer);
}

CLP_FFI_GO_METHOD auto ir_serializer_eight_byte_create(void** ir_serializer_ptr, ByteSpan* ir_span)
        -> int {
    return serializer_create<eight_byte_encoded_variable_t>(*ir_serializer_ptr, ir_span);
}

CLP_FFI_GO_METHOD auto ir_serializer_four_byte_create(void** ir_serializer_ptr, ByteSpan* ir_span)
        -> int {
    return serializer_create<four_byte_encoded_variable_t>(*ir_serializer_ptr, ir_span);
}

CLP_FFI_GO_METHOD auto ir_serializer_eight_byte_serialize_log_event(
        void* ir_serializer,
        ByteSpan auto_kv_msgpack_bytes,
        ByteSpan user_kv_msgpack_bytes,
        ByteSpan* ir_span
) -> int {
    return serialize_log_event<eight_byte_encoded_variable_t>(
            ir_serializer,
            auto_kv_msgpack_bytes,
            user_kv_msgpack_bytes,
            ir_span
    );
}

CLP_FFI_GO_METHOD auto ir_serializer_four_byte_serialize_log_event(
        void* ir_serializer,
        ByteSpan auto_kv_msgpack_bytes,
        ByteSpan user_kv_msgpack_bytes,
        ByteSpan* ir_span
) -> int {
    return serialize_log_event<four_byte_encoded_variable_t>(
            ir_serializer,
            auto_kv_msgpack_bytes,
            user_kv_msgpack_bytes,
            ir_span
    );
}
}  // namespace clp_ffi_go::ir
