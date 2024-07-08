#include "deserializer.h"

#include <algorithm>
#include <cstddef>
#include <cstdint>
#include <functional>
#include <span>
#include <string_view>
#include <type_traits>
#include <utility>
#include <vector>

#include <clp/BufferReader.hpp>
#include <clp/ErrorCode.hpp>
#include <clp/ffi/ir_stream/decoding_methods.hpp>
#include <clp/ffi/ir_stream/protocol_constants.hpp>
#include <clp/ir/types.hpp>
#include <clp/string_utils/string_utils.hpp>
#include <clp/time_types.hpp>

#include "ffi_go/api_decoration.h"
#include "ffi_go/defs.h"
#include "ffi_go/ir/types.hpp"
#include "ffi_go/search/wildcard_query.h"
#include "ffi_go/types.hpp"

namespace ffi_go::ir {
using clp::BufferReader;
using clp::ffi::ir_stream::cProtocol::Eof;
using clp::ffi::ir_stream::cProtocol::Payload::UtcOffsetChange;
using clp::ffi::ir_stream::deserialize_preamble;
using clp::ffi::ir_stream::deserialize_tag;
using clp::ffi::ir_stream::encoded_tag_t;
using clp::ffi::ir_stream::get_encoding_type;
using clp::ffi::ir_stream::IRErrorCode;
using clp::ir::eight_byte_encoded_variable_t;
using clp::ir::four_byte_encoded_variable_t;
using clp::UtcOffset;

namespace {
/**
 * Generic helper for ir_deserializer_deserialize_*_log_event
 */
template <class encoded_variable_t>
[[nodiscard]] auto deserialize_log_event(
        ByteSpan ir_view,
        void* ir_deserializer,
        size_t* ir_pos,
        LogEventView* log_event
) -> int;

/**
 * Generic helper for ir_deserializer_deserialize_*_wildcard_match
 */
template <class encoded_variable_t>
[[nodiscard]] auto deserialize_wildcard_match(
        ByteSpan ir_view,
        void* ir_deserializer,
        TimestampInterval time_interval,
        WildcardQueryView queries,
        size_t* ir_pos,
        LogEventView* log_event,
        size_t* matching_query
) -> int;

/**
 * Deserializes UTC offset changes until the next read tag is not a UTC offset change packet.
 * @param ir_buf
 * @param tag Outputs the tag after deserializing UTC offset changes.
 * @param utc_offset Outputs the last deserialized UTC offset.
 * @return IRErrorCode::IRErrorCode_Success on success.
 * @return IRErrorCode::IRErrorCode_Incomplete_IR if the reader doesn't contain enough data to
 * deserialize.
 */
[[nodiscard]] auto deserialize_utc_offset_changes(
        BufferReader& ir_buf,
        encoded_tag_t& tag,
        UtcOffset& utc_offset
) -> IRErrorCode;

template <class encoded_variable_t>
auto deserialize_log_event(
        ByteSpan ir_view,
        void* ir_deserializer,
        size_t* ir_pos,
        LogEventView* log_event
) -> int {
    if (nullptr == ir_deserializer || nullptr == ir_pos || nullptr == log_event) {
        return static_cast<int>(IRErrorCode::IRErrorCode_Corrupted_IR);
    }
    BufferReader ir_buf{static_cast<char const*>(ir_view.m_data), ir_view.m_size};
    Deserializer* deserializer{static_cast<Deserializer*>(ir_deserializer)};

    encoded_tag_t tag{};
    if (auto const err{deserialize_tag(ir_buf, tag)}; IRErrorCode::IRErrorCode_Success != err) {
        return static_cast<int>(err);
    }
    if (UtcOffsetChange == tag) {
        UtcOffset utc_offset{0};
        if (auto const err{deserialize_utc_offset_changes(ir_buf, tag, utc_offset)};
            IRErrorCode::IRErrorCode_Success != err)
        {
            return err;
        }
        deserializer->m_utc_offset = utc_offset;
    }
    if (Eof == tag) {
        return static_cast<int>(IRErrorCode::IRErrorCode_Eof);
    }

    IRErrorCode err{};
    epoch_time_ms_t timestamp{};
    if constexpr (std::is_same_v<encoded_variable_t, eight_byte_encoded_variable_t>) {
        err = clp::ffi::ir_stream::eight_byte_encoding::deserialize_log_event(
                ir_buf,
                tag,
                deserializer->m_log_event.m_log_message,
                timestamp
        );
    } else if constexpr (std::is_same_v<encoded_variable_t, four_byte_encoded_variable_t>) {
        epoch_time_ms_t timestamp_delta{};
        err = clp::ffi::ir_stream::four_byte_encoding::deserialize_log_event(
                ir_buf,
                tag,
                deserializer->m_log_event.m_log_message,
                timestamp_delta
        );
        timestamp = deserializer->m_timestamp + timestamp_delta;
    } else {
        static_assert(cAlwaysFalse<encoded_variable_t>, "Invalid/unhandled encoding type");
    }
    if (IRErrorCode::IRErrorCode_Success != err) {
        return static_cast<int>(err);
    }
    deserializer->m_timestamp = timestamp;

    size_t pos{0};
    if (clp::ErrorCode_Success != ir_buf.try_get_pos(pos)) {
        return static_cast<int>(IRErrorCode::IRErrorCode_Decode_Error);
    }
    *ir_pos = pos;
    log_event->m_log_message.m_data = deserializer->m_log_event.m_log_message.data();
    log_event->m_log_message.m_size = deserializer->m_log_event.m_log_message.size();
    log_event->m_timestamp = deserializer->m_timestamp;
    log_event->m_utc_offset = static_cast<epoch_time_ms_t>(deserializer->m_utc_offset.count());
    return static_cast<int>(IRErrorCode::IRErrorCode_Success);
}

template <class encoded_variable_t>
auto deserialize_wildcard_match(
        ByteSpan ir_view,
        void* ir_deserializer,
        TimestampInterval time_interval,
        MergedWildcardQueryView merged_query,
        size_t* ir_pos,
        LogEventView* log_event,
        size_t* matching_query
) -> int {
    if (nullptr == ir_deserializer || nullptr == ir_pos || nullptr == log_event
        || nullptr == matching_query)
    {
        return static_cast<int>(IRErrorCode::IRErrorCode_Corrupted_IR);
    }
    BufferReader ir_buf{static_cast<char const*>(ir_view.m_data), ir_view.m_size};
    Deserializer* deserializer{static_cast<Deserializer*>(ir_deserializer)};
    std::string_view const query_view{merged_query.m_queries.m_data, merged_query.m_queries.m_size};
    std::span<size_t> const end_offsets{
            merged_query.m_end_offsets.m_data,
            merged_query.m_end_offsets.m_size
    };
    std::span<bool> const case_sensitivity{
            merged_query.m_case_sensitivity.m_data,
            merged_query.m_case_sensitivity.m_size
    };

    std::vector<std::pair<std::string_view, bool>> queries(merged_query.m_end_offsets.m_size);
    size_t pos{0};
    for (size_t i{0}; i < merged_query.m_end_offsets.m_size; i++) {
        auto& [query_str_view, is_case_sensitive]{queries[i]};
        query_str_view = query_view.substr(pos, end_offsets[i]);
        is_case_sensitive = case_sensitivity[i];
        pos += end_offsets[i];
    }

    std::function<std::pair<bool, size_t>(ffi_go::LogMessage const&)> query_fn;
    if (false == queries.empty()) {
        query_fn = [&](ffi_go::LogMessage const& log_message) -> std::pair<bool, size_t> {
            auto const found_query = std::find_if(
                    queries.cbegin(),
                    queries.cend(),
                    [&](std::pair<std::string_view, bool> const& query) -> bool {
                        return clp::string_utils::wildcard_match_unsafe(
                                log_message,
                                query.first,
                                query.second
                        );
                    }
            );
            return {queries.cend() != found_query, found_query - queries.cbegin()};
        };
    } else {
        query_fn = [](ffi_go::LogMessage const&) -> std::pair<bool, size_t> { return {true, 0}; };
    }

    IRErrorCode err{};
    while (true) {
        encoded_tag_t tag{};
        if (err = deserialize_tag(ir_buf, tag); IRErrorCode::IRErrorCode_Success != err) {
            return static_cast<int>(err);
        }
        if (UtcOffsetChange == tag) {
            UtcOffset utc_offset{0};
            if (err = deserialize_utc_offset_changes(ir_buf, tag, utc_offset);
                IRErrorCode::IRErrorCode_Success != err)
            {
                return err;
            }
            deserializer->m_utc_offset = utc_offset;
        }
        if (Eof == tag) {
            return static_cast<int>(IRErrorCode::IRErrorCode_Eof);
        }

        epoch_time_ms_t timestamp{};
        if constexpr (std::is_same_v<encoded_variable_t, eight_byte_encoded_variable_t>) {
            err = clp::ffi::ir_stream::eight_byte_encoding::deserialize_log_event(
                    ir_buf,
                    tag,
                    deserializer->m_log_event.m_log_message,
                    timestamp
            );
        } else if constexpr (std::is_same_v<encoded_variable_t, four_byte_encoded_variable_t>) {
            epoch_time_ms_t timestamp_delta{};
            err = clp::ffi::ir_stream::four_byte_encoding::deserialize_log_event(
                    ir_buf,
                    tag,
                    deserializer->m_log_event.m_log_message,
                    timestamp_delta
            );
            timestamp = deserializer->m_timestamp + timestamp_delta;
        } else {
            static_assert(cAlwaysFalse<encoded_variable_t>, "Invalid/unhandled encoding type");
        }
        if (IRErrorCode::IRErrorCode_Success != err) {
            return static_cast<int>(err);
        }
        deserializer->m_timestamp = timestamp;

        if (time_interval.m_upper <= deserializer->m_timestamp) {
            // TODO this is an extremely fragile hack until the CLP ffi ir
            // code is refactored and IRErrorCode includes things beyond
            // decoding.
            return static_cast<int>(IRErrorCode::IRErrorCode_Incomplete_IR + 1);
        }
        if (time_interval.m_lower > deserializer->m_timestamp) {
            continue;
        }
        auto const [has_matching_query, matching_query_idx]{
                query_fn(deserializer->m_log_event.m_log_message)
        };
        if (false == has_matching_query) {
            continue;
        }
        size_t curr_ir_pos{0};
        if (clp::ErrorCode_Success != ir_buf.try_get_pos(curr_ir_pos)) {
            return static_cast<int>(IRErrorCode::IRErrorCode_Decode_Error);
        }
        *ir_pos = curr_ir_pos;
        log_event->m_log_message.m_data = deserializer->m_log_event.m_log_message.data();
        log_event->m_log_message.m_size = deserializer->m_log_event.m_log_message.size();
        log_event->m_timestamp = deserializer->m_timestamp;
        log_event->m_utc_offset = static_cast<epoch_time_ms_t>(deserializer->m_utc_offset.count());
        *matching_query = matching_query_idx;
        return static_cast<int>(IRErrorCode::IRErrorCode_Success);
    }
}

auto deserialize_utc_offset_changes(BufferReader& ir_buf, encoded_tag_t& tag, UtcOffset& utc_offset)
        -> IRErrorCode {
    while (UtcOffsetChange == tag) {
        if (auto const err{clp::ffi::ir_stream::deserialize_utc_offset_change(ir_buf, utc_offset)})
        {
            return err;
        }
        if (auto const err{deserialize_tag(ir_buf, tag)}) {
            return err;
        }
    }
    return IRErrorCode::IRErrorCode_Success;
}
}  // namespace

CLP_FFI_GO_METHOD auto ir_deserializer_close(void* ir_deserializer) -> void {
    // NOLINTNEXTLINE(cppcoreguidelines-owning-memory)
    delete static_cast<Deserializer*>(ir_deserializer);
}

CLP_FFI_GO_METHOD auto ir_deserializer_new_deserializer_with_preamble(
        ByteSpan ir_view,
        size_t* ir_pos,
        int8_t* ir_encoding,
        int8_t* metadata_type,
        size_t* metadata_pos,
        uint16_t* metadata_size,
        void** ir_deserializer_ptr,
        void** timestamp_ptr
) -> int {
    if (nullptr == ir_pos || nullptr == ir_encoding || nullptr == metadata_type
        || nullptr == metadata_pos || nullptr == metadata_size || nullptr == ir_deserializer_ptr
        || nullptr == timestamp_ptr)
    {
        return static_cast<int>(IRErrorCode::IRErrorCode_Corrupted_IR);
    }
    BufferReader ir_buf{static_cast<char const*>(ir_view.m_data), ir_view.m_size};

    bool four_byte_encoding{};
    if (IRErrorCode const err{get_encoding_type(ir_buf, four_byte_encoding)};
        IRErrorCode::IRErrorCode_Success != err)
    {
        return static_cast<int>(err);
    }
    *ir_encoding = four_byte_encoding ? 1 : 0;

    if (IRErrorCode const err{
                deserialize_preamble(ir_buf, *metadata_type, *metadata_pos, *metadata_size)
        };
        IRErrorCode::IRErrorCode_Success != err)
    {
        return static_cast<int>(err);
    }

    size_t pos{0};
    if (clp::ErrorCode_Success != ir_buf.try_get_pos(pos)) {
        return static_cast<int>(IRErrorCode::IRErrorCode_Decode_Error);
    }
    *ir_pos = pos;
    auto* deserializer{new Deserializer()};
    *ir_deserializer_ptr = deserializer;
    *timestamp_ptr = &deserializer->m_timestamp;
    return static_cast<int>(IRErrorCode::IRErrorCode_Success);
}

CLP_FFI_GO_METHOD auto ir_deserializer_deserialize_eight_byte_log_event(
        ByteSpan ir_view,
        void* ir_deserializer,
        size_t* ir_pos,
        LogEventView* log_event
) -> int {
    return deserialize_log_event<eight_byte_encoded_variable_t>(
            ir_view,
            ir_deserializer,
            ir_pos,
            log_event
    );
}

CLP_FFI_GO_METHOD auto ir_deserializer_deserialize_four_byte_log_event(
        ByteSpan ir_view,
        void* ir_deserializer,
        size_t* ir_pos,
        LogEventView* log_event
) -> int {
    return deserialize_log_event<four_byte_encoded_variable_t>(
            ir_view,
            ir_deserializer,
            ir_pos,
            log_event
    );
}

CLP_FFI_GO_METHOD auto ir_deserializer_deserialize_eight_byte_wildcard_match(
        ByteSpan ir_view,
        void* ir_deserializer,
        TimestampInterval time_interval,
        MergedWildcardQueryView merged_query,
        size_t* ir_pos,
        LogEventView* log_event,
        size_t* matching_query
) -> int {
    return deserialize_wildcard_match<eight_byte_encoded_variable_t>(
            ir_view,
            ir_deserializer,
            time_interval,
            merged_query,
            ir_pos,
            log_event,
            matching_query
    );
}

CLP_FFI_GO_METHOD auto ir_deserializer_deserialize_four_byte_wildcard_match(
        ByteSpan ir_view,
        void* ir_deserializer,
        TimestampInterval time_interval,
        MergedWildcardQueryView merged_query,
        size_t* ir_pos,
        LogEventView* log_event,
        size_t* matching_query
) -> int {
    return deserialize_wildcard_match<four_byte_encoded_variable_t>(
            ir_view,
            ir_deserializer,
            time_interval,
            merged_query,
            ir_pos,
            log_event,
            matching_query
    );
}
}  // namespace ffi_go::ir
