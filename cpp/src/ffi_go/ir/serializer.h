#ifndef FFI_GO_IR_SERIALIZER_H
#define FFI_GO_IR_SERIALIZER_H
// header must support C, making modernize checks inapplicable
// NOLINTBEGIN(modernize-deprecated-headers)
// NOLINTBEGIN(modernize-use-trailing-return-type)

#include <stdint.h>
#include <stdlib.h>

#include <ffi_go/defs.h>

/**
 * Clean up the underlying ir::Serializer of a Go ir.Serializer.
 * @param[in] ir_serializer Address of a ir::Serializer created and returned by
 *     ir_serializer_serialize_*_preamble
 */
void ir_serializer_close(void* ir_serializer);

/**
 * Given the fields of a CLP IR preamble, serialize them into an IR byte stream
 * with eight byte encoding. An ir::Serializer will be allocated to use as the
 * backing storage for a Go ir.Serializer (i.e. subsequent calls to
 * ir_serializer_serialize_*_log_event). All pointer parameters must be non-null
 * (non-nil Cgo C.<type> pointer or unsafe.Pointer from Go).
 * @param[in] ts_pattern Format string for the timestamp to be used when
 *     deserializing the IR
 * @param[in] ts_pattern_syntax Type of the format string for understanding how
 *     to parse it
 * @param[in] time_zone_id TZID timezone of the timestamps in the IR
 * @param[out] ir_serializer_ptr Address of a new ir::Serializer
 * @param[out] ir_view View of a IR buffer containing the serialized preamble
 * @return ffi::ir_stream::IRErrorCode forwarded from
 *     ffi::ir_stream::eight_byte_encoding::encode_preamble
 */
int ir_serializer_serialize_eight_byte_preamble(
        StringView ts_pattern,
        StringView ts_pattern_syntax,
        StringView time_zone_id,
        void** ir_serializer_ptr,
        ByteSpan* ir_view
);

/**
 * Given the fields of a CLP IR preamble, serialize them into an IR byte stream
 * with four byte encoding. An ir::Serializer will be allocated to use as the
 * backing storage for a Go ir.Serializer (i.e. subsequent calls to
 * ir_serializer_serialize_*_log_event). All pointer parameters must be non-null
 * (non-nil Cgo C.<type> pointer or unsafe.Pointer from Go).
 * @param[in] ts_pattern Format string for the timestamp to be used when
 *     deserializing the IR
 * @param[in] ts_pattern_syntax Type of the format string for understanding how
 *     to parse it
 * @param[in] time_zone_id TZID timezone of the timestamps in the IR
 * @param[out] ir_serializer_ptr Address of a new ir::Serializer
 * @param[out] ir_view View of a IR buffer containing the serialized preamble
 * @return ffi::ir_stream::IRErrorCode forwarded from
 *     ffi::ir_stream::four_byte_encoding::encode_preamble
 */
int ir_serializer_serialize_four_byte_preamble(
        StringView ts_pattern,
        StringView ts_pattern_syntax,
        StringView time_zone_id,
        epoch_time_ms_t reference_ts,
        void** ir_serializer_ptr,
        ByteSpan* ir_view
);

/**
 * Given the fields of a log event, serialize them into an IR byte stream with
 * eight byte encoding. An ir::Serializer must be provided to use as the backing
 * storage for the corresponding Go ir.Serializer. All pointer parameters must
 * be non-null (non-nil Cgo C.<type> pointer or unsafe.Pointer from Go).
 * @param[in] log_message Log message of the log event to serialize
 * @param[in] timestamp Timestamp of the log event to serialize
 * @param[in] ir_serializer ir::Serializer object to be used as storage
 * @param[out] ir_view View of a IR buffer containing the serialized log event
 * @return ffi::ir_stream::IRErrorCode forwarded from
 *   ffi::ir_stream::eight_byte_encoding::encode_message
 */
int ir_serializer_serialize_eight_byte_log_event(
        StringView log_message,
        epoch_time_ms_t timestamp,
        void* ir_serializer,
        ByteSpan* ir_view
);

/**
 * Given the fields of a log event, serialize them into an IR byte stream with
 * four byte encoding. An ir::Serializer must be provided to use as the backing
 * storage for the corresponding Go ir.Serializer. All pointer parameters must
 * be non-null (non-nil Cgo C.<type> pointer or unsafe.Pointer from Go).
 * @param[in] log_message Log message to serialize
 * @param[in] timestamp_delta Timestamp delta to the previous log event in the
 *     IR stream
 * @param[in] ir_serializer ir::Serializer object to be used as storage
 * @param[out] ir_view View of a IR buffer containing the serialized log event
 * @return ffi::ir_stream::IRErrorCode forwarded from
 *     ffi::ir_stream::four_byte_encoding::encode_message
 */
int ir_serializer_serialize_four_byte_log_event(
        StringView log_message,
        epoch_time_ms_t timestamp_delta,
        void* ir_serializer,
        ByteSpan* ir_view
);

// NOLINTEND(modernize-use-trailing-return-type)
// NOLINTEND(modernize-deprecated-headers)
#endif  // FFI_GO_IR_SERIALIZER_H
