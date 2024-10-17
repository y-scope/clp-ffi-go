#include "serializer.h"

#include <memory>
#include <lint/msgpack.hpp>
#include <boost/outcome.hpp>
#include <system_error>

#include <clp/ffi/ir_stream/Serializer.hpp>
#include <clp/ir/types.hpp>

#include "ffi_go/api_decoration.h"
#include "ffi_go/defs.h"
#include "ffi_go/ir/types.hpp"

namespace ffi_go::ir {
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
[[nodiscard]] auto serializer_create(
        void*& ir_serializer_ptr,
        ByteSpan* ir_view
) -> int;

/**
 * Generic helper for ir_serializer_serialize_*_log_event functions.
 */
template <class encoded_variable_t>
[[nodiscard]] auto
serialize_log_event(void* ir_serializer, ByteSpan msgpack_bytes, ByteSpan* ir_view) -> int;

template <class encoded_variable_t>
auto serializer_close(void* ir_serializer) -> void {
    std::unique_ptr<clp::ffi::ir_stream::Serializer<encoded_variable_t>>(static_cast<clp::ffi::ir_stream::Serializer<encoded_variable_t>*>(ir_serializer));
}

template <class encoded_variable_t>
auto serializer_create(
        void*& ir_serializer_ptr,
        ByteSpan* ir_view
) -> int {
    if (nullptr == ir_serializer_ptr || nullptr == ir_view) {
        return static_cast<int>(std::errc::protocol_error);
    }
    auto result{clp::ffi::ir_stream::Serializer<encoded_variable_t>::create()};
    if (result.has_failure()) {
        return result.error().value();
    }
    auto ir_buf_view{result.value().get_ir_buf_view()};
    ir_view->m_data = ir_buf_view.data();
    ir_view->m_size = ir_buf_view.size();
    auto s = std::make_unique<clp::ffi::ir_stream::Serializer<encoded_variable_t>>(
            std::move(result.value())
    );
    ir_serializer_ptr = s.release();
    return 0;
}

template <class encoded_variable_t>
auto serialize_log_event(void* ir_serializer, ByteSpan msgpack_bytes, ByteSpan* ir_view) -> int {
    if (nullptr == ir_serializer || nullptr == ir_view) {
        return static_cast<int>(std::errc::protocol_error);
    }
    auto* serializer{static_cast<clp::ffi::ir_stream::Serializer<encoded_variable_t>*>(ir_serializer
    )};

    auto const mp_handle{
            msgpack::unpack(static_cast<char const*>(msgpack_bytes.m_data), msgpack_bytes.m_size)
    };
    /* if (serializer->serialize_msgpack_map(mp_handle.get())) { */
    if (serializer->serialize_msgpack_map(mp_handle.get().via.map)) {
        return static_cast<int>(std::errc::protocol_error);
    }

    auto ir_buf_view{serializer->get_ir_buf_view()};
    ir_view->m_data = ir_buf_view.data();
    ir_view->m_size = ir_buf_view.size();
    return 0;
}
}  // namespace

CLP_FFI_GO_METHOD auto ir_serializer_eight_byte_close(void* ir_serializer) -> void {
    serializer_close<eight_byte_encoded_variable_t>(ir_serializer);
}

CLP_FFI_GO_METHOD auto ir_serializer_four_byte_close(void* ir_serializer) -> void {
    serializer_close<four_byte_encoded_variable_t>(ir_serializer);
}

CLP_FFI_GO_METHOD auto ir_serializer_eight_byte_create(
        void** ir_serializer_ptr,
        ByteSpan* ir_view
) -> int {
    return serializer_create<eight_byte_encoded_variable_t>(
            *ir_serializer_ptr,
            ir_view
    );
}

CLP_FFI_GO_METHOD auto ir_serializer_four_byte_create(
        void** ir_serializer_ptr,
        ByteSpan* ir_view
) -> int {
    return serializer_create<four_byte_encoded_variable_t>(
            *ir_serializer_ptr,
            ir_view
    );
}

CLP_FFI_GO_METHOD auto ir_serializer_eight_byte_serialize_log_event(
        void* ir_serializer,
        ByteSpan msgpack_bytes,
        ByteSpan* ir_view
) -> int {
    return serialize_log_event<eight_byte_encoded_variable_t>(
            ir_serializer,
            msgpack_bytes,
            ir_view
    );
}

CLP_FFI_GO_METHOD auto ir_serializer_four_byte_serialize_log_event(
        void* ir_serializer,
        ByteSpan msgpack_bytes,
        ByteSpan* ir_view
) -> int {
    return serialize_log_event<four_byte_encoded_variable_t>(
            ir_serializer,
            msgpack_bytes,
            ir_view
    );
}
}  // namespace ffi_go::ir
