#include "deserializer.h"

#include <cstddef>
#include <cstdint>
#include <ffi/ir_stream/IrUnitType.hpp>
#include <memory>
#include <system_error>
#include <utility>
#include <vector>

#include <clp_ffi_go/api_decoration.h>
#include <clp_ffi_go/defs.h>
#include <clp_ffi_go/ir/ErrorCode.hpp>

#include <clp/BufferReader.hpp>
#include <clp/ErrorCode.hpp>
#include <clp/ffi/ir_stream/decoding_methods.hpp>
#include <clp/ffi/ir_stream/Deserializer.hpp>
#include <clp/ffi/KeyValuePairLogEvent.hpp>
#include <clp/ffi/SchemaTree.hpp>
#include <clp/time_types.hpp>
#include <nlohmann/json.hpp>
#include <outcome/outcome.hpp>

namespace clp_ffi_go::ir {
using clp::ffi::ir_stream::IRErrorCode;
using clp::ffi::ir_stream::IrUnitType;

namespace {
/**
 * Implements `clp::ffi::ir_stream::IrUnitHandlerInterface` interface
 */
class IrUnitHandler {
public:
    [[nodiscard]] auto handle_log_event([[maybe_unused]] clp::ffi::KeyValuePairLogEvent&& log_event)
            -> IRErrorCode {
        auto result{log_event.serialize_to_json()};
        if (result.has_failure()) {
            return IRErrorCode::IRErrorCode_Corrupted_IR;
        }
        m_auto_kv_pairs = nlohmann::json::to_msgpack(result.value().first);
        m_user_kv_pairs = nlohmann::json::to_msgpack(result.value().second);
        return IRErrorCode::IRErrorCode_Success;
    }

    [[nodiscard]] static auto handle_utc_offset_change(
            [[maybe_unused]] clp::UtcOffset utc_offset_old,
            [[maybe_unused]] clp::UtcOffset utc_offset_new
    ) -> IRErrorCode {
        return IRErrorCode::IRErrorCode_Success;
    }

    [[nodiscard]] static auto handle_schema_tree_node_insertion(
            [[maybe_unused]] bool is_auto_generated,
            [[maybe_unused]] clp::ffi::SchemaTree::NodeLocator schema_tree_node_locator,
            [[maybe_unused]] std::shared_ptr<clp::ffi::SchemaTree const> schema_tree
    ) -> IRErrorCode {
        return IRErrorCode::IRErrorCode_Success;
    }

    [[nodiscard]] auto handle_end_of_stream() -> IRErrorCode {
        m_is_complete = true;
        return IRErrorCode::IRErrorCode_Success;
    }

    // Methods
    [[nodiscard]] auto is_complete() const -> bool { return m_is_complete; }

    [[nodiscard]] auto get_auto_kv_pairs() const -> std::vector<uint8_t> const& {
        return m_auto_kv_pairs;
    }

    [[nodiscard]] auto get_user_kv_pairs() const -> std::vector<uint8_t> const& {
        return m_user_kv_pairs;
    }

private:
    std::vector<uint8_t> m_auto_kv_pairs;
    std::vector<uint8_t> m_user_kv_pairs;
    bool m_is_complete{false};
};
}  // namespace

CLP_FFI_GO_METHOD auto ir_deserializer_close(void* ir_deserializer) -> void {
    std::unique_ptr<clp::ffi::ir_stream::Deserializer<IrUnitHandler>>(
            static_cast<clp::ffi::ir_stream::Deserializer<IrUnitHandler>*>(ir_deserializer)
    );
}

CLP_FFI_GO_METHOD auto
ir_deserializer_create(ByteSpan ir_span, size_t* ir_pos, void** ir_deserializer_ptr) -> int {
    if (nullptr == ir_pos || nullptr == ir_deserializer_ptr) {
        return static_cast<int>(ErrorCode::GoIrError::InvalidArg);
    }

    clp::BufferReader ir_buf{static_cast<char const*>(ir_span.m_data), ir_span.m_size};
    auto deserializer_result{
            clp::ffi::ir_stream::Deserializer<IrUnitHandler>::create(ir_buf, IrUnitHandler{})
    };
    if (deserializer_result.has_failure()) {
        return static_cast<int>(ErrorCode::convert_to_go_ir_error(deserializer_result.error()));
    }

    size_t pos{0};
    if (clp::ErrorCode_Success != ir_buf.try_get_pos(pos)) {
        return static_cast<int>(ErrorCode::GoIrError::IrBufferError);
    }
    *ir_pos = pos;
    auto d = std::make_unique<clp::ffi::ir_stream::Deserializer<IrUnitHandler>>(
            std::move(deserializer_result.value())
    );
    *ir_deserializer_ptr = d.release();
    return static_cast<int>(ErrorCode::GoIrError::Success);
}

CLP_FFI_GO_METHOD auto ir_deserializer_deserialize_log_event(
        ByteSpan ir_span,
        void* ir_deserializer,
        size_t* ir_pos,
        ByteSpan* auto_kv_pairs_span,
        ByteSpan* user_kv_pairs_span
) -> int {
    if (nullptr == ir_deserializer || nullptr == ir_pos || nullptr == auto_kv_pairs_span
        || nullptr == user_kv_pairs_span)
    {
        return static_cast<int>(ErrorCode::GoIrError::InvalidArg);
    }
    clp::BufferReader ir_reader{static_cast<char const*>(ir_span.m_data), ir_span.m_size};
    auto* deserializer{
            static_cast<clp::ffi::ir_stream::Deserializer<IrUnitHandler>*>(ir_deserializer)
    };

    while (true) {
        auto result{deserializer->deserialize_next_ir_unit(ir_reader)};
        if (result.has_failure()) {
            if (result.error() == std::errc::result_out_of_range) {
                return static_cast<int>(ErrorCode::GoIrError::IrIncomplete);
            }
            return static_cast<int>(ErrorCode::convert_to_go_ir_error(result.error()));
        }
        // Update the buffer position for Go on each successful IR unit
        size_t pos{0};
        if (clp::ErrorCode_Success != ir_reader.try_get_pos(pos)) {
            return static_cast<int>(ErrorCode::GoIrError::IrBufferError);
        }
        *ir_pos = pos;
        switch (result.value()) {
            case IrUnitType::LogEvent: {
                auto const& auto_kv_pairs{deserializer->get_ir_unit_handler().get_auto_kv_pairs()};
                auto_kv_pairs_span->m_data = auto_kv_pairs.data();
                auto_kv_pairs_span->m_size = auto_kv_pairs.size();
                auto const& user_kv_pairs{deserializer->get_ir_unit_handler().get_user_kv_pairs()};
                user_kv_pairs_span->m_data = user_kv_pairs.data();
                user_kv_pairs_span->m_size = user_kv_pairs.size();
                return static_cast<int>(ErrorCode::GoIrError::Success);
            }
            case IrUnitType::EndOfStream: {
                return static_cast<int>(ErrorCode::GoIrError::IrEndOfStream);
            }
            case IrUnitType::SchemaTreeNodeInsertion:
            case IrUnitType::UtcOffsetChange: {
                continue;
            }
            default:
                return static_cast<int>(ErrorCode::GoIrError::NotSupported);
        }
    }
}
}  // namespace clp_ffi_go::ir
