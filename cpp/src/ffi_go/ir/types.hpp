#ifndef FFI_GO_IR_LOG_TYPES_HPP
#define FFI_GO_IR_LOG_TYPES_HPP

#include <cstddef>
#include <cstdint>
#include <string>
#include <vector>

#include <clp/ir/types.hpp>

#include "ffi_go/types.hpp"

namespace ffi_go::ir {

template <typename>
[[maybe_unused]] constexpr bool cAlwaysFalse{false};

template <typename encoded_var_t>
struct LogMessage {
    auto reserve(size_t cap) -> void { m_logtype.reserve(cap); }

    std::string m_logtype;
    std::vector<encoded_var_t> m_vars;
    std::vector<char> m_dict_vars;
    std::vector<int32_t> m_dict_var_end_offsets;
};

/**
 * The backing storage for a Go ir.Decoder.
 * Mutating a field will invalidate the corresponding View (slice) stored in the
 * ir.Decoder (without any warning or way to guard in Go).
 */
struct Decoder {
    ffi_go::LogMessage m_log_message;
};

/**
 * The backing storage for a Go ir.Encoder.
 * Mutating a field will invalidate the corresponding View (slice) stored in the
 * ir.Encoder (without any warning or way to guard in Go).
 */
template <typename encoded_var_t>
struct Encoder {
    LogMessage<encoded_var_t> m_log_message;
};

/**
 * The backing storage for a Go ir.Deserializer.
 * Mutating a field will invalidate the corresponding View (slice) stored in the
 * ir.Deserializer (without any warning or way to guard in Go).
 */
struct Deserializer {
    ffi_go::LogEventStorage m_log_event;
    clp::ir::epoch_time_ms_t m_timestamp{};
};

/**
 * The backing storage for a Go ir.Serializer.
 * Mutating a field will invalidate the corresponding View (slice) stored in the
 * ir.Serializer (without any warning or way to guard in Go).
 */
struct Serializer {
    /**
     * Reserve capacity for the logtype and ir buffer.
     * We reserve 1.5x the size of the log message type as a heuristic for the
     * full IR buffer size. The log message type of a log event is not
     * guaranteed to be less than or equal to the size of the actual log
     * message, but in general this is true.
     */
    auto reserve(size_t cap) -> void {
        m_logtype.reserve(cap);
        m_ir_buf.reserve(cap + cap / 2);
    }

    std::string m_logtype;
    std::vector<int8_t> m_ir_buf;
};
}  // namespace ffi_go::ir

#endif  // FFI_GO_IR_LOG_TYPES_HPP
