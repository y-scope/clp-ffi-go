#include "deserializer.h"

#include <algorithm>
#include <cstddef>
#include <cstdint>
#include <functional>
#include <span>
#include <string_view>
#include <type_traits>
#include <utility>
#include <vector>

#include <clp/BufferReader.hpp>
#include <clp/ErrorCode.hpp>
#include <clp/time_types.hpp>
#include <clp/ffi/KeyValuePairLogEvent.hpp>
#include <clp/ffi/ir_stream/Deserializer.hpp>
#include <clp/ffi/ir_stream/decoding_methods.hpp>
#include <clp/ffi/ir_stream/protocol_constants.hpp>
#include <clp/ir/types.hpp>
#include <clp/string_utils/string_utils.hpp>

#include "ffi_go/api_decoration.h"
#include "ffi_go/defs.h"
#include "ffi_go/ir/types.hpp"
#include "ffi_go/types.hpp"

namespace ffi_go::ir {
using clp::ffi::ir_stream::cProtocol::Eof;
using clp::ffi::ir_stream::deserialize_tag;
using clp::ffi::ir_stream::get_encoding_type;
using clp::ffi::ir_stream::IRErrorCode;
using clp::ir::eight_byte_encoded_variable_t;
using clp::ir::four_byte_encoded_variable_t;

namespace {
/**
 * Class that implements `clp::ffi::ir_stream::IrUnitHandlerInterface` for testing purposes.
 */
class IrUnitHandler {
public:
    // Implements `clp::ffi::ir_stream::IrUnitHandlerInterface` interface
    [[nodiscard]] auto handle_log_event([[maybe_unused]] clp::ffi::KeyValuePairLogEvent&& log_event) -> IRErrorCode {
        return IRErrorCode::IRErrorCode_Success;
    }

    [[nodiscard]] static auto handle_utc_offset_change(
            [[maybe_unused]] clp::UtcOffset utc_offset_old,
            [[maybe_unused]] clp::UtcOffset utc_offset_new
    ) -> IRErrorCode {
        return IRErrorCode::IRErrorCode_Success;
    }

    [[nodiscard]] static auto handle_schema_tree_node_insertion(
            [[maybe_unused]] clp::ffi::SchemaTree::NodeLocator schema_tree_node_locator
    ) -> IRErrorCode {
        return IRErrorCode::IRErrorCode_Success;
    }

    [[nodiscard]] auto handle_end_of_stream() -> IRErrorCode {
        m_is_complete = true;
        return IRErrorCode::IRErrorCode_Success;
    }

    // Methods
    [[nodiscard]] auto is_complete() const -> bool { return m_is_complete; }

    [[nodiscard]] auto get_msgpack_log_event() const -> std::vector<int8_t> const& {
        return m_msgpack_log_event;
    }

private:
    std::vector<int8_t> m_msgpack_log_event;
    bool m_is_complete{false};
};

/*/1** */
/* * Generic helper for ir_deserializer_*_close functions. */
/* *1/ */
/*template <class encoded_variable_t> */
/*auto deserializer_close(void* ir_deserializer) -> void; */

/*/1** */
/* * Generic helper for ir_deserializer_deserialize_*_log_event */
/* *1/ */
/*template <class encoded_variable_t> */
/*[[nodiscard]] auto deserialize_log_event( */
/*        ByteSpan ir_view, */
/*        void* ir_deserializer, */
/*        size_t* ir_pos, */
/*        LogEventView* log_event */
/*) -> int; */

/*template <class encoded_variable_t> */
/*auto deserializer_close(void* ir_deserializer) -> void { */
/*    std::unique_ptr<clp::ffi::ir_stream::Deserializer<IrUnitHandler>>(static_cast<clp::ffi::ir_stream::Deserializer<IrUnitHandler>*>(ir_deserializer)); */
/*} */

/* template <class encoded_variable_t> */
/* auto deserialize_log_event( */
/*         ByteSpan ir_view, */
/*         void* ir_deserializer, */
/*         size_t* ir_pos, */
/*         LogEventView* log_event */
/* ) -> int { */
/*     if (nullptr == ir_deserializer || nullptr == ir_pos || nullptr == log_event) { */
/*         return static_cast<int>(IRErrorCode::IRErrorCode_Corrupted_IR); */
/*     } */
/*     clp::BufferReader ir_buf{static_cast<char const*>(ir_view.m_data), ir_view.m_size}; */
/*     auto deserializer{static_cast<clp::ffi::ir_stream::Deserializer<encoded_variable_t>*>(ir_deserializer)}; */

/*     clp::ffi::ir_stream::encoded_tag_t tag{}; */
/*     if (auto const err{deserialize_tag(ir_buf, tag)}; IRErrorCode::IRErrorCode_Success != err) { */
/*         return static_cast<int>(err); */
/*     } */
/*     if (Eof == tag) { */
/*         return static_cast<int>(IRErrorCode::IRErrorCode_Eof); */
/*     } */

/*     IRErrorCode err{}; */
/*     epoch_time_ms_t timestamp{}; */
/*     if constexpr (std::is_same_v<encoded_variable_t, eight_byte_encoded_variable_t>) { */
/*         err = clp::ffi::ir_stream::eight_byte_encoding::deserialize_log_event( */
/*                 ir_buf, */
/*                 tag, */
/*                 deserializer->m_log_event.m_log_message, */
/*                 timestamp */
/*         ); */
/*     } else if constexpr (std::is_same_v<encoded_variable_t, four_byte_encoded_variable_t>) { */
/*         epoch_time_ms_t timestamp_delta{}; */
/*         err = clp::ffi::ir_stream::four_byte_encoding::deserialize_log_event( */
/*                 ir_buf, */
/*                 tag, */
/*                 deserializer->m_log_event.m_log_message, */
/*                 timestamp_delta */
/*         ); */
/*         timestamp = deserializer->m_timestamp + timestamp_delta; */
/*     } else { */
/*         static_assert(cAlwaysFalse<encoded_variable_t>, "Invalid/unhandled encoding type"); */
/*     } */
/*     if (IRErrorCode::IRErrorCode_Success != err) { */
/*         return static_cast<int>(err); */
/*     } */
/*     deserializer->m_timestamp = timestamp; */

/*     size_t pos{0}; */
/*     if (clp::ErrorCode_Success != ir_buf.try_get_pos(pos)) { */
/*         return static_cast<int>(IRErrorCode::IRErrorCode_Decode_Error); */
/*     } */
/*     *ir_pos = pos; */
/*     log_event->m_log_message.m_data = deserializer->m_log_event.m_log_message.data(); */
/*     log_event->m_log_message.m_size = deserializer->m_log_event.m_log_message.size(); */
/*     log_event->m_timestamp = deserializer->m_timestamp; */
/*     return static_cast<int>(IRErrorCode::IRErrorCode_Success); */
/* } */
}  // namespace

/* CLP_FFI_GO_METHOD auto ir_deserializer_eight_byte_close(void* ir_deserializer) -> void { */
/*     deserializer_close<eight_byte_encoded_variable_t>(ir_deserializer); */
/* } */

/* CLP_FFI_GO_METHOD auto ir_deserializer_four_byte_close(void* ir_deserializer) -> void { */
/*     deserializer_close<four_byte_encoded_variable_t>(ir_deserializer); */
/* } */

CLP_FFI_GO_METHOD auto ir_deserializer_close(void* ir_deserializer) -> void {
    std::unique_ptr<clp::ffi::ir_stream::Deserializer<IrUnitHandler>>(static_cast<clp::ffi::ir_stream::Deserializer<IrUnitHandler>*>(ir_deserializer));
}

CLP_FFI_GO_METHOD auto ir_deserializer_create(
        ByteSpan ir_view,
        size_t* ir_pos,
        void** ir_deserializer_ptr
) -> int {
    if (nullptr == ir_pos || nullptr == ir_deserializer_ptr) {
        return static_cast<int>(std::errc::protocol_error);
    }

    clp::BufferReader ir_buf{static_cast<char const*>(ir_view.m_data), ir_view.m_size};
    auto deserializer_result{clp::ffi::ir_stream::Deserializer<IrUnitHandler>::create(ir_buf, IrUnitHandler{})};
    if (deserializer_result.has_failure()) {
        return deserializer_result.error().value();
    }

    size_t pos{0};
    if (clp::ErrorCode_Success != ir_buf.try_get_pos(pos)) {
        return static_cast<int>(std::errc::protocol_error);
    }
    *ir_pos = pos;
    auto d = std::make_unique<clp::ffi::ir_stream::Deserializer<IrUnitHandler>>(std::move(deserializer_result.value()));
    *ir_deserializer_ptr = d.release();
    return 0;
}

/* CLP_FFI_GO_METHOD auto ir_deserializer_deserialize_eight_byte_log_event( */
/*         ByteSpan ir_view, */
/*         void* ir_deserializer, */
/*         size_t* ir_pos, */
/*         LogEventView* log_event */
/* ) -> int { */
/*     return deserialize_log_event<eight_byte_encoded_variable_t>( */
/*             ir_view, */
/*             ir_deserializer, */
/*             ir_pos, */
/*             log_event */
/*     ); */
/* } */

/* CLP_FFI_GO_METHOD auto ir_deserializer_deserialize_four_byte_log_event( */
/*         ByteSpan ir_view, */
/*         void* ir_deserializer, */
/*         size_t* ir_pos, */
/*         LogEventView* log_event */
/* ) -> int { */
/*     return deserialize_log_event<four_byte_encoded_variable_t>( */
/*             ir_view, */
/*             ir_deserializer, */
/*             ir_pos, */
/*             log_event */
/*     ); */
/* } */

CLP_FFI_GO_METHOD auto ir_deserializer_deserialize_log_event(
        [[maybe_unused]] ByteSpan ir_view,
        [[maybe_unused]] void* ir_deserializer,
        [[maybe_unused]] size_t* ir_pos,
        [[maybe_unused]] ByteSpan* msgpack_log_event_view
) -> int {
    return 0;
}
}  // namespace ffi_go::ir
