#ifndef CLP_FFI_GO_IR_ERROR_HPP
#define CLP_FFI_GO_IR_ERROR_HPP

#include <cstdint>
#include <system_error>

#include <clp/ffi/ir_stream/decoding_methods.hpp>

namespace clp_ffi_go::ir {
/*
 * TODO
 */
class ErrorCode {
public:
    /*
     * Mirrors the Go type ir.IrError (ir/IrError.go) and is used when returning through cgo. The
     * values of these types must match.
     */
    enum class GoIrError : uint8_t {
        Success = 0,
        InvalidArg,
        IrBufferError,
        IrCorrupt,
        IrDecodeError,
        IrEndOfStream,
        IrIncomplete,
        IrProtoBackwardCompatible,
        IrProtoSupported,
        IrProtoUnsupported,
        IrSerializeError,
        NotSupported,
    };

    /**
     * @param ir_error_code
     * @return Equivalent `IrError` (ir/irerror.go) int value indicating the same error.
     */
    [[nodiscard]] static auto convert_to_go_ir_error(clp::ffi::ir_stream::IRErrorCode code)
            -> GoIrError;

    /**
     * @param ir_error_code
     * @return Equivalent `IrError` (ir/irerror.go) int value indicating the same error.
     */
    [[nodiscard]] static auto convert_to_go_ir_error(clp::ffi::ir_stream::IRProtocolErrorCode code)
            -> GoIrError;

    /**
     * @param err
     * @return Equivalent `IrError` (ir/irerror.go) int value indicating the same error.
     */
    [[nodiscard]] static auto convert_to_go_ir_error(std::error_code code) -> GoIrError;
};
}  // namespace clp_ffi_go::ir
#endif  // CLP_FFI_GO_IR_ERROR_HPP
