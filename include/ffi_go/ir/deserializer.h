#ifndef FFI_GO_IR_DESERIALIZER_H
#define FFI_GO_IR_DESERIALIZER_H
// header must support C, making modernize checks inapplicable
// NOLINTBEGIN(modernize-deprecated-headers)
// NOLINTBEGIN(modernize-use-trailing-return-type)

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>
#include <stdlib.h>

#include <ffi_go/defs.h>
#include <ffi_go/search/wildcard_query.h>

    /**
     * Clean up the underlying ir::Deserializer of a Go ir.Deserializer.
     * @param[in] ir_deserializer The address of a ir::Deserializer created and
     *     returned by ir_deserializer_deserialize_preamble
     */
    void ir_deserializer_close(void* ir_deserializer);

    /**
     * Given a CLP IR buffer (any encoding), attempt to deserialize a premable and
     * extract its information. An ir::Deserializer will be allocated to use as the
     * backing storage for a Go ir.Deserializer (i.e. subsequent calls to
     * ir_deserializer_deserialize_*_log_event). It is left to the Go layer to read
     * the metadata based on the returned type. All pointer parameters must be
     * non-null (non-nil Cgo C.<type> pointer or unsafe.Pointer from Go).
     * @param[in] ir_view Byte buffer/slice containing CLP IR
     * @param[out] ir_pos Position in ir_view read to
     * @param[out] ir_encoding IR encoding type (1: four byte, 0: eight byte)
     * @param[out] metadata_type Type of metadata in preamble (e.g. json)
     * @param[out] metadata_pos Position in ir_view where the metadata begins
     * @param[out] metadata_size Size of the metadata (in bytes)
     * @param[out] ir_deserializer_ptr Address of a new ir::Deserializer
     * @param[out] timestamp_ptr Address of m_timestamp inside the ir::Deserializer
     *     to be filled in by Go using the metadata contents
     * @return ffi::ir_stream::IRErrorCode forwarded from either
     *     ffi::ir_stream::get_encoding_type or ffi::ir_stream::decode_preamble
     */
    int ir_deserializer_deserialize_preamble(
            ByteSpan ir_view,
            size_t* ir_pos,
            int8_t* ir_encoding,
            int8_t* metadata_type,
            size_t* metadata_pos,
            uint16_t* metadata_size,
            void** ir_deserializer_ptr,
            void** timestamp_ptr
    );

    /**
     * Given a CLP IR buffer with eight byte encoding, deserialize the next log
     * event. Returns the components of the found log event and the buffer position
     * it ends at. All pointer parameters must be non-null (non-nil Cgo C.<type>
     * pointer or unsafe.Pointer from Go).
     * @param[in] ir_view Byte buffer/slice containing CLP IR
     * @param[in] ir_deserializer ir::Deserializer to be used as storage for a found
     *     log event
     * @param[out] ir_pos Position in ir_view read to
     * @param[out] log_event Log event stored in ir_deserializer
     * @return ffi::ir_stream::IRErrorCode forwarded from
     *     ffi::ir_stream::eight_byte_encoding::decode_next_message
     */
    int ir_deserializer_deserialize_eight_byte_log_event(
            ByteSpan ir_view,
            void* ir_deserializer,
            size_t* ir_pos,
            LogEventView* log_event
    );

    /**
     * Given a CLP IR buffer with four byte encoding, deserialize the next log
     * event. Returns the components of the found log event and the buffer position
     * it ends at. All pointer parameters must be non-null (non-nil Cgo C.<type>
     * pointer or unsafe.Pointer from Go).
     * @param[in] ir_view Byte buffer/slice containing CLP IR
     * @param[in] ir_deserializer ir::Deserializer to be used as storage for a found
     *     log event
     * @param[out] ir_pos Position in ir_view read to
     * @param[out] log_event Log event stored in ir_deserializer
     * @return ffi::ir_stream::IRErrorCode forwarded from
     *     ffi::ir_stream::four_byte_encoding::decode_next_message
     */
    int ir_deserializer_deserialize_four_byte_log_event(
            ByteSpan ir_view,
            void* ir_deserializer,
            size_t* ir_pos,
            LogEventView* log_event
    );

    /**
     * Given a CLP IR buffer with eight byte encoding, deserialize the next log
     * event until finding an event that is both within the time interval and
     * matches any query. If queries is empty, the first log event within the time
     * interval is treated as a match. Returns the components of the found log event
     * and the buffer position it ends at. All pointer parameters must be non-null
     * (non-nil Cgo C.<type> pointer or unsafe.Pointer from Go).
     * @param[in] ir_view Byte buffer/slice containing CLP IR
     * @param[in] ir_deserializer ir::Deserializer to be used as storage for a found
     *     log event
     * @param[in] time_interval Timestamp interval: [lower, upper)
     * @param[in] merged_query A concatenation of all queries to filter for; if
     *     empty any log event as a match
     * @param[out] ir_pos Position in ir_view read to
     * @param[out] log_event Log event stored in ir_deserializer
     * @param[out] matching_query Index into queries of the first matching query or
     *     0 if queries is empty
     * @return ffi::ir_stream::IRErrorCode forwarded from
     *     ffi::ir_stream::four_byte_encoding::decode_next_message
     * @return ffi::ir_stream::IRErrorCode_Unsupported_Version + 1 if no query is
     *     found before time_interval.m_upper (TODO this should be replaced/fix in
     *     clp core)
     */
    int ir_deserializer_deserialize_eight_byte_wildcard_match(
            ByteSpan ir_view,
            void* ir_deserializer,
            TimestampInterval time_interval,
            MergedWildcardQueryView merged_query,
            size_t* ir_pos,
            LogEventView* log_event,
            size_t* matching_query
    );

    /**
     * Given a CLP IR buffer with four byte encoding, deserialize the next log event
     * until finding an event that is both within the time interval and matches any
     * query. If queries is empty, the first log event within the time interval is
     * treated as a match. Returns the components of the found log event and the
     * buffer position it ends at. All pointer parameters must be non-null (non-nil
     * Cgo C.<type> pointer or unsafe.Pointer from Go).
     * @param[in] ir_view Byte buffer/slice containing CLP IR
     * @param[in] ir_deserializer ir::Deserializer to be used as storage for a found
     *     log event
     * @param[in] time_interval Timestamp interval: [lower, upper)
     * @param[in] merged_query A concatenation of all queries to filter for; if
     *     empty any log event as a match
     * @param[out] ir_pos Position in ir_view read to
     * @param[out] log_event Log event stored in ir_deserializer
     * @param[out] matching_query Index into queries of the matching query
     * @return ffi::ir_stream::IRErrorCode forwarded from
     *     ffi::ir_stream::four_byte_encoding::decode_next_message
     * @return ffi::ir_stream::IRErrorCode_Unsupported_Version + 1 if no query is
     *     found before time_interval.m_upper (TODO this should be replaced/fix in
     *     clp core)
     */
    int ir_deserializer_deserialize_four_byte_wildcard_match(
            ByteSpan ir_view,
            void* ir_deserializer,
            TimestampInterval time_interval,
            MergedWildcardQueryView merged_query,
            size_t* ir_pos,
            LogEventView* log_event,
            size_t* matching_query
    );

#ifdef __cplusplus
}
#endif

// NOLINTEND(modernize-use-trailing-return-type)
// NOLINTEND(modernize-deprecated-headers)
#endif  // FFI_GO_IR_DESERIALIZER_H
