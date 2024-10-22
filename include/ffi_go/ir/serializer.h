#ifndef FFI_GO_IR_SERIALIZER_H
#define FFI_GO_IR_SERIALIZER_H
// header must support C, making modernize checks inapplicable
// NOLINTBEGIN(modernize-use-trailing-return-type)

#include "ffi_go/api_decoration.h"
#include "ffi_go/defs.h"

/**
 * Clean up the underlying `clp::ffi::ir_stream::Serializer` of a Go `ir.Serializer`.
 * @param[in] ir_serializer Address of a `clp::ffi::ir_stream::Serializer` created and returned by
 *     `ir_serializer_eight_byte_create`
 */
CLP_FFI_GO_METHOD void ir_serializer_eight_byte_close(void* ir_serializer);

/**
 * Clean up the underlying `clp::ffi::ir_stream::Serializer` of a Go `ir.Serializer`.
 * @param[in] ir_serializer Address of a `clp::ffi::ir_stream::Serializer` created and returned by
 *     `ir_serializer_four_byte_create`
 */
CLP_FFI_GO_METHOD void ir_serializer_four_byte_close(void* ir_serializer);

/**
 * Given the fields of a CLP IR preamble, serialize them into an IR byte stream with eight byte
 * encoding. A `clp::ffi::ir_stream::Serializer` will be allocated to use as the backing storage for
 * a Go `ir.Serializer` (i.e. subsequent calls to `ir_serializer_serialize_eight_byte_*`).
 * @param[out] ir_serializer_ptr Address of a new `clp::ffi::ir_stream::Serializer`.
 * @param[out] ir_view View of a IR buffer containing the serialized preamble.
 * @return 0 on success.
 * @return `std::errc::protocol_error` value on null arguments.
 * @return Forward's `clp::ffi::ir_stream::Serializer::create` return values.
 */
CLP_FFI_GO_METHOD int
ir_serializer_eight_byte_create(void** ir_serializer_ptr, ByteSpan* ir_view);

/**
 * Given the fields of a CLP IR preamble, serialize them into an IR byte stream with four byte
 * encoding. A `clp::ffi::ir_stream::Serializer` will be allocated to use as the backing storage for
 * a Go `ir.Serializer` (i.e. subsequent calls to `ir_serializer_serialize_four_byte_*`).
 * @param[out] ir_serializer_ptr Address of a new `clp::ffi::ir_stream::Serializer`.
 * @param[out] ir_view View of a IR buffer containing the serialized preamble.
 * @return 0 on success.
 * @return `std::errc::protocol_error` value on null arguments.
 * @return Forward's `clp::ffi::ir_stream::Serializer::create` return values.
 */
CLP_FFI_GO_METHOD int
ir_serializer_four_byte_create(void** ir_serializer_ptr, ByteSpan* ir_view);

/**
 * Given a log event encoded as a msgpack map, serialize it into an IR byte stream with eight byte
 * encoding. A `clp::ffi::ir_stream::Serializer` must be provided to use as the backing storage for
 * the corresponding Go `ir.Serializer`.
 * @param[in] ir_serializer_ptr `clp::ffi::ir_stream::Serializer` object to be used as storage.
 * @param[in] msgpack_bytes log event encoded as a msgpack map.
 * @param[out] ir_view View of a IR buffer containing the serialized log event.
 * @return 0 on success.
 * @return `std::errc::protocol_error` value on null arguments.
 * @return Forward's `clp::ffi::ir_stream::Serializer::serialize_msgpack_map` return values.
 */
CLP_FFI_GO_METHOD int ir_serializer_eight_byte_serialize_log_event(
        void* ir_serializer,
        ByteSpan msgpack_bytes,
        ByteSpan* ir_view
);

/**
 * Given a log event encoded as a msgpack map, serialize it into an IR byte stream with four byte
 * encoding. A `clp::ffi::ir_stream::Serializer` must be provided to use as the backing storage for
 * the corresponding Go `ir.Serializer`.
 * @param[in] ir_serializer_ptr `clp::ffi::ir_stream::Serializer` object to be used as storage.
 * @param[in] msgpack_bytes log event encoded as a msgpack map.
 * @param[out] ir_view View of a IR buffer containing the serialized log event.
 * @return 0 on success.
 * @return `std::errc::protocol_error` value on null arguments.
 * @return Forward's `clp::ffi::ir_stream::Serializer::serialize_msgpack_map` return values.
 */
CLP_FFI_GO_METHOD int ir_serializer_four_byte_serialize_log_event(
        void* ir_serializer,
        ByteSpan msgpack_bytes,
        ByteSpan* ir_view
);

// NOLINTEND(modernize-use-trailing-return-type)
#endif  // FFI_GO_IR_SERIALIZER_H
