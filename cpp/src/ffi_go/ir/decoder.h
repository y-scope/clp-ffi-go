#ifndef FFI_GO_IR_DECODER_H
#define FFI_GO_IR_DECODER_H
// header must support C, making modernize checks inapplicable
// NOLINTBEGIN(modernize-deprecated-headers)
// NOLINTBEGIN(modernize-use-trailing-return-type)
// NOLINTBEGIN(modernize-use-using)

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>
#include <stdlib.h>

#include <ffi_go/defs.h>

    /**
     * Create a ir::Decoder used as the underlying data storage for a Go ir.Decoder.
     * @return New ir::Decoder's address
     */
    void* ir_decoder_new();

    /**
     * Clean up the underlying ir::Decoder of a Go ir.Decoder.
     * @param[in] ir_encoder Address of a ir::Decoder created and returned by
     *   ir_decoder_new
     */
    void ir_decoder_close(void* decoder);

    /**
     * Given the fields of a CLP IR encoded log message with eight byte encoding,
     * decode it into the original log message. An ir::Decoder must be provided to
     * use as the backing storage for the corresponding Go ir.Decoder. All pointer
     * parameters must be non-null (non-nil Cgo C.<type> pointer or unsafe.Pointer
     * from Go).
     * @param[in] logtype Type of the log message (the log message with variables
     *     extracted and replaced with placeholders)
     * @param[in] vars Array of encoded variables
     * @param[in] dict_vars String containing all dictionary variables concatenated
     *     together
     * @param[in] dict_var_end_offsets Array of offsets into dict_vars marking the
     *     end of a dictionary variable
     * @param[in] ir_decoder ir::Decoder to be used as storage for the decoded log
     *     message
     * @param[out] log_message Decoded log message
     * @return ffi::ir_stream::IRErrorCode_Decode_Error if ffi::decode_message
     *     throws or errors
     * @return ffi::ir_stream::IRErrorCode_Success on success
     */
    int ir_decoder_decode_eight_byte_log_message(
            StringView logtype,
            Int64tSpan vars,
            StringView dict_vars,
            Int32tSpan dict_var_end_offsets,
            void* ir_decoder,
            StringView* log_message
    );

    /**
     * Given the fields of a CLP IR encoded log message with four byte encoding,
     * decode it into the original log message. An ir::Decoder must be provided to
     * use as the backing storage for the corresponding Go ir.Decoder. All pointer
     * parameters must be non-null (non-nil Cgo C.<type> pointer or unsafe.Pointer
     * from Go).
     * @param[in] logtype Type of the log message (the log message with variables
     *     extracted and replaced with placeholders)
     * @param[in] vars Array of encoded variables
     * @param[in] dict_vars String containing all dictionary variables concatenated
     *     together
     * @param[in] dict_var_end_offsets Array of offsets into dict_vars makring the
     *     end of a dictionary variable
     * @param[in] ir_decoder ir::Decoder to be used as storage for the decoded log
     *     message
     * @param[out] log_message Decoded log message
     * @return ffi::ir_stream::IRErrorCode_Decode_Error if ffi::decode_message
     *     throws or errors
     * @return ffi::ir_stream::IRErrorCode_Success on success
     */
    int ir_decoder_decode_four_byte_log_message(
            StringView logtype,
            Int32tSpan vars,
            StringView dict_vars,
            Int32tSpan dict_var_end_offsets,
            void* ir_decoder,
            StringView* log_message
    );

#ifdef __cplusplus
}
#endif

// NOLINTEND(modernize-use-using)
// NOLINTEND(modernize-use-trailing-return-type)
// NOLINTEND(modernize-deprecated-headers)
#endif  // FFI_GO_IR_DECODER_H
