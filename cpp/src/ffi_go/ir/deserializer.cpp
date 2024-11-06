#include "deserializer.h"

#include <cstddef>
#include <cstdint>
#include <ffi/ir_stream/IrUnitType.hpp>
#include <json/single_include/nlohmann/json.hpp>
#include <memory>
#include <outcome/single-header/outcome.hpp>
#include <system_error>
#include <utility>
#include <vector>

#include <clp/BufferReader.hpp>
#include <clp/ErrorCode.hpp>
#include <clp/ffi/ir_stream/decoding_methods.hpp>
#include <clp/ffi/ir_stream/Deserializer.hpp>
#include <clp/ffi/KeyValuePairLogEvent.hpp>
#include <clp/ffi/SchemaTree.hpp>
#include <clp/time_types.hpp>

#include "ffi_go/api_decoration.h"
#include "ffi_go/defs.h"

namespace ffi_go::ir {
using clp::ffi::ir_stream::IRErrorCode;
using clp::ffi::ir_stream::IrUnitType;

namespace {
/**
 * Implements `clp::ffi::ir_stream::IrUnitHandlerInterface` interface
 */
class IrUnitHandler {
public:
    [[nodiscard]] auto handle_log_event([[maybe_unused]] clp::ffi::KeyValuePairLogEvent&& log_event
    ) -> IRErrorCode {
        auto result{log_event.serialize_to_json()};
        if (result.has_failure()) {
            /* return result.error().value(); */
            return IRErrorCode::IRErrorCode_Corrupted_IR;
        }
        m_msgpack_log_event = nlohmann::json::to_msgpack(result.value());
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

    [[nodiscard]] auto get_msgpack_log_event() const -> std::vector<uint8_t> const& {
        return m_msgpack_log_event;
    }

private:
    std::vector<uint8_t> m_msgpack_log_event;
    bool m_is_complete{false};
};
}  // namespace

CLP_FFI_GO_METHOD auto ir_deserializer_close(void* ir_deserializer) -> void {
    std::unique_ptr<clp::ffi::ir_stream::Deserializer<IrUnitHandler>>(
            static_cast<clp::ffi::ir_stream::Deserializer<IrUnitHandler>*>(ir_deserializer)
    );
}

CLP_FFI_GO_METHOD auto
ir_deserializer_create(ByteSpan ir_view, size_t* ir_pos, void** ir_deserializer_ptr) -> int {
    if (nullptr == ir_pos || nullptr == ir_deserializer_ptr) {
        return static_cast<int>(std::errc::protocol_error);
    }

    clp::BufferReader ir_buf{static_cast<char const*>(ir_view.m_data), ir_view.m_size};
    auto deserializer_result{
            clp::ffi::ir_stream::Deserializer<IrUnitHandler>::create(ir_buf, IrUnitHandler{})
    };
    if (deserializer_result.has_failure()) {
        return deserializer_result.error().value();
    }

    size_t pos{0};
    if (clp::ErrorCode_Success != ir_buf.try_get_pos(pos)) {
        return static_cast<int>(std::errc::protocol_error);
    }
    *ir_pos = pos;
    auto d = std::make_unique<clp::ffi::ir_stream::Deserializer<IrUnitHandler>>(
            std::move(deserializer_result.value())
    );
    *ir_deserializer_ptr = d.release();
    return 0;
}

CLP_FFI_GO_METHOD auto ir_deserializer_deserialize_log_event(
        ByteSpan ir_view,
        void* ir_deserializer,
        size_t* ir_pos,
        ByteSpan* msgpack_log_event_view
) -> int {
    if (nullptr == ir_deserializer || nullptr == ir_pos || nullptr == msgpack_log_event_view) {
        return static_cast<int>(std::errc::protocol_error);
    }
    clp::BufferReader ir_reader{static_cast<char const*>(ir_view.m_data), ir_view.m_size};
    auto* deserializer{
            static_cast<clp::ffi::ir_stream::Deserializer<IrUnitHandler>*>(ir_deserializer)
    };

    while (true) {
        auto result{deserializer->deserialize_next_ir_unit(ir_reader)};
        if (result.has_failure()) {
            if (result.error() == std::errc::result_out_of_range) {
                return IRErrorCode::IRErrorCode_Incomplete_IR;
            }
            /* return result.error().value(); */
            return IRErrorCode::IRErrorCode_Corrupted_IR;
        }
        // Update the buffer position for Go on each successful IR unit
        size_t pos{0};
        if (clp::ErrorCode_Success != ir_reader.try_get_pos(pos)) {
            return static_cast<int>(IRErrorCode::IRErrorCode_Decode_Error);
        }
        *ir_pos = pos;
        switch (result.value()) {
            case IrUnitType::LogEvent: {
                auto const& msgpack_buf{deserializer->get_ir_unit_handler().get_msgpack_log_event()
                };
                msgpack_log_event_view->m_data = msgpack_buf.data();
                msgpack_log_event_view->m_size = msgpack_buf.size();
                return IRErrorCode::IRErrorCode_Success;
            }
            case IrUnitType::EndOfStream: {
                return static_cast<int>(IRErrorCode::IRErrorCode_Eof);
            }
            case IrUnitType::SchemaTreeNodeInsertion:
            case IrUnitType::UtcOffsetChange: {
                continue;
            }
            default:
                /* return std::errc::protocol_not_supported; */
                return IRErrorCode::IRErrorCode_Corrupted_IR;
        }
    }
}
}  // namespace ffi_go::ir
