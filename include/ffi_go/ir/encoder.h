#ifndef FFI_GO_IR_ENCODER_H
#define FFI_GO_IR_ENCODER_H
// header must support C, making modernize checks inapplicable
// NOLINTBEGIN(modernize-use-trailing-return-type)

#include "ffi_go/api_decoration.h"
#include "ffi_go/defs.h"

/**
 * Create a ir::Encoder used as the underlying data storage for a Go ir.Encoder.
 * @return New ir::Encoder's address
 */
CLP_FFI_GO_METHOD void* ir_encoder_eight_byte_new();

/**
 * @copydoc ir_encoder_eight_byte_new()
 */
CLP_FFI_GO_METHOD void* ir_encoder_four_byte_new();

/**
 * Clean up the underlying ir::Encoder of a Go ir.Encoder.
 * @param[in] ir_encoder Address of a ir::Encoder created and returned by
 *   ir_encoder_eight_byte_new
 */
CLP_FFI_GO_METHOD void ir_encoder_eight_byte_close(void* ir_encoder);

/**
 * Clean up the underlying ir::Encoder of a Go ir.Encoder.
 * @param[in] ir_encoder Address of a ir::Encoder created and returned by
 *   ir_encoder_four_byte_new
 */
CLP_FFI_GO_METHOD void ir_encoder_four_byte_close(void* ir_encoder);

/**
 * Given a log message, encode it into a CLP IR object with eight byte encoding.
 * An ir::Encoder must be provided to use as the backing storage for the
 * corresponding Go ir.Encoder. All pointer parameters must be non-null (non-nil
 * Cgo C.<type> pointer or unsafe.Pointer from Go).
 * @param[in] log_message Log message to encode
 * @param[in] ir_encoder ir::Encoder to be used as storage for the encoded log
 *     message
 * @param[out] logtype Type of the log message (the log message with variables
 *     extracted and replaced with placeholders)
 * @param[out] vars Array of encoded variables
 * @param[out] dict_vars String containing all dictionary variables concatenated
 *     together
 * @param[out] dict_var_end_offsets Array of offsets into dict_vars marking the
 *     end of a dictionary variable
 * @return ffi::ir_stream::IRErrorCode_Corrupted_IR if ffi::encode_message
 *   returns false
 * @return ffi::ir_stream::IRErrorCode_Success on success
 */
CLP_FFI_GO_METHOD int ir_encoder_encode_eight_byte_log_message(
        StringView log_message,
        void* ir_encoder,
        StringView* logtype,
        Int64tSpan* vars,
        StringView* dict_vars,
        Int32tSpan* dict_var_end_offsets
);

/**
 * Given a log message, encode it into a CLP IR object with four byte encoding.
 * An ir::Encoder must be provided to use as the backing storage for the
 * corresponding Go ir.Encoder. All pointer parameters must be non-null (non-nil
 * Cgo C.<type> pointer or unsafe.Pointer from Go).
 * @param[in] log_message Log message to encode
 * @param[in] ir_encoder ir::Encoder to be used as storage for the encoded log
 *     message
 * @param[out] logtype Type of the log message (the log message with variables
 *     extracted and replaced with placeholders)
 * @param[out] vars Array of encoded variables
 * @param[out] dict_vars String containing all dictionary variables concatenated
 *     together
 * @param[out] dict_var_end_offsets Array of offsets into dict_vars marking the
 *     end of a dictionary variable
 * @return ffi::ir_stream::IRErrorCode_Corrupted_IR if ffi::encode_message
 *   returns false
 * @return ffi::ir_stream::IRErrorCode_Success on success
 */
CLP_FFI_GO_METHOD int ir_encoder_encode_four_byte_log_message(
        StringView log_message,
        void* ir_encoder,
        StringView* logtype,
        Int32tSpan* vars,
        StringView* dict_vars,
        Int32tSpan* dict_var_end_offsets
);

// NOLINTEND(modernize-use-trailing-return-type)
#endif  // FFI_GO_IR_ENCODER_H
