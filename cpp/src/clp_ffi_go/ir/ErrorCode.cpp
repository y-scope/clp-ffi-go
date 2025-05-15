#include "ErrorCode.hpp"

#include <system_error>

#include <clp/ffi/ir_stream/decoding_methods.hpp>

namespace clp_ffi_go::ir {
using clp::ffi::ir_stream::IRErrorCode;
using clp::ffi::ir_stream::IRProtocolErrorCode;

// This conversion must be kept up to date with the IRErrorCode definition in:
// clp/components/core/src/ffi/ir_stream/decoding_methods.hpp
[[nodiscard]] auto ErrorCode::convert_to_go_ir_error(IRErrorCode code) -> GoIrError {
    switch (code) {
        case IRErrorCode::IRErrorCode_Success:
            return GoIrError::Success;
        case IRErrorCode::IRErrorCode_Decode_Error:
            return GoIrError::IrDecodeError;
        case IRErrorCode::IRErrorCode_Eof:
            return GoIrError::IrEndOfStream;
        case IRErrorCode::IRErrorCode_Corrupted_IR:
            return GoIrError::IrCorrupt;
        case IRErrorCode::IRErrorCode_Incomplete_IR:
            return GoIrError::IrIncomplete;
        default:
            return GoIrError::NotSupported;
    }
}

// This conversion must be kept up to date with the IRErrorCode definition in:
// clp/components/core/src/ffi/ir_stream/decoding_methods.hpp
[[nodiscard]] auto ErrorCode::convert_to_go_ir_error(IRProtocolErrorCode code) -> GoIrError {
    switch (code) {
        case IRProtocolErrorCode::Supported:
            return GoIrError::IrProtoSupported;
        case IRProtocolErrorCode::BackwardCompatible:
            return GoIrError::IrProtoBackwardCompatible;
        case IRProtocolErrorCode::Unsupported:
            return GoIrError::IrProtoUnsupported;
        case IRProtocolErrorCode::Invalid:
        default:
            return GoIrError::NotSupported;
    }
}

// This conversion must be kept up to date with the IRErrorCode to std::errc conversion found in:
// components/core/src/clp/ffi/ir_stream/utils.cpp.
[[nodiscard]] auto ErrorCode::convert_to_go_ir_error(std::error_code code) -> GoIrError {
    if (std::errc::result_out_of_range == code) {
        return GoIrError::IrIncomplete;
    }
    if (std::errc::result_out_of_range == code) {
        return GoIrError::IrIncomplete;
    }
    if (std::errc::protocol_error == code) {
        return GoIrError::IrCorrupt;
    }
    if (std::errc::no_message_available == code) {
        return GoIrError::IrEndOfStream;
    }
    return GoIrError::NotSupported;
}
}  // namespace clp_ffi_go::ir
