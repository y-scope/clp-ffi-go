#ifndef FFI_GO_LOG_TYPES_HPP
#define FFI_GO_LOG_TYPES_HPP

#include <string>

namespace ffi_go {
/**
 * The backing storage for a Go ffi.LogMessageView.
 * Mutating it will invalidate the corresponding View (slice) stored in the
 * ffi.LogMessageView (without any warning or way to guard in Go).
 */
using LogMessage = std::string;

/**
 * The backing storage for a Go ffi.LogEventView.
 * Mutating a field will invalidate the corresponding View (slice) stored in the
 * ffi.LogEventView (without any warning or way to guard in Go).
 */
struct LogEvent {
    auto reserve(size_t cap) -> void { m_log_message.reserve(cap); }

    LogMessage m_log_message;
};
}  // namespace ffi_go

#endif  // FFI_GO_LOG_TYPES_HPP
