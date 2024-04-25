#include "deserializer.h"

#include <algorithm>
#include <cstddef>
#include <cstdint>
#include <functional>
#include <memory>
#include <span>
#include <string>
#include <string_view>
#include <type_traits>

#include <clp/components/core/src/BufferReader.hpp>
#include <clp/components/core/src/ErrorCode.hpp>
#include <clp/components/core/src/ffi/encoding_methods.hpp>
#include <clp/components/core/src/ffi/ir_stream/decoding_methods.hpp>
#include <clp/components/core/src/ffi/ir_stream/encoding_methods.hpp>

#include <ffi_go/defs.h>
#include <ffi_go/ir/LogTypes.hpp>
#include <ffi_go/LogTypes.hpp>
#include <ffi_go/search/wildcard_query.h>

namespace ffi_go::ir {
using namespace ffi;
using namespace ffi::ir_stream;

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

    template <class encoded_variable_t>
    auto deserialize_log_event(
            ByteSpan ir_view,
            void* ir_deserializer,
            size_t* ir_pos,
            LogEventView* log_event
    ) -> int {
        BufferReader ir_buf{static_cast<char const*>(ir_view.m_data), ir_view.m_size};
        Deserializer* deserializer{static_cast<Deserializer*>(ir_deserializer)};

        IRErrorCode err{};
        epoch_time_ms_t timestamp{};
        if constexpr (std::is_same_v<encoded_variable_t, eight_byte_encoded_variable_t>) {
            err = eight_byte_encoding::decode_next_message(
                    ir_buf,
                    deserializer->m_log_event.m_log_message,
                    timestamp
            );
        } else if constexpr (std::is_same_v<encoded_variable_t, four_byte_encoded_variable_t>) {
            epoch_time_ms_t timestamp_delta{};
            err = four_byte_encoding::decode_next_message(
                    ir_buf,
                    deserializer->m_log_event.m_log_message,
                    timestamp_delta
            );
            timestamp = deserializer->m_timestamp + timestamp_delta;
        } else {
            static_assert(cAlwaysFalse<encoded_variable_t>, "Invalid/unhandled encoding type");
        }
        if (IRErrorCode_Success != err) {
            return static_cast<int>(err);
        }
        deserializer->m_timestamp = timestamp;

        size_t pos{0};
        if (ErrorCode_Success != ir_buf.try_get_pos(pos)) {
            return static_cast<int>(IRErrorCode_Decode_Error);
        }
        *ir_pos = pos;
        log_event->m_log_message.m_data = deserializer->m_log_event.m_log_message.data();
        log_event->m_log_message.m_size = deserializer->m_log_event.m_log_message.size();
        log_event->m_timestamp = deserializer->m_timestamp;
        return static_cast<int>(IRErrorCode_Success);
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
        BufferReader ir_buf{static_cast<char const*>(ir_view.m_data), ir_view.m_size};
        Deserializer* deserializer{static_cast<Deserializer*>(ir_deserializer)};
        std::string_view const query_view{
                merged_query.m_queries.m_data,
                merged_query.m_queries.m_size};
        std::span<size_t> const end_offsets{
                merged_query.m_end_offsets.m_data,
                merged_query.m_end_offsets.m_size};
        std::span<bool> const case_sensitivity{
                merged_query.m_case_sensitivity.m_data,
                merged_query.m_case_sensitivity.m_size};

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
                            return wildcard_match_unsafe(log_message, query.first, query.second);
                        }
                );
                return {queries.cend() != found_query, found_query - queries.cbegin()};
            };
        } else {
            query_fn = [](ffi_go::LogMessage const&) -> std::pair<bool, size_t> {
                return {true, 0};
            };
        }

        IRErrorCode err{};
        while (true) {
            epoch_time_ms_t timestamp{};
            if constexpr (std::is_same_v<encoded_variable_t, eight_byte_encoded_variable_t>) {
                err = eight_byte_encoding::decode_next_message(
                        ir_buf,
                        deserializer->m_log_event.m_log_message,
                        timestamp
                );
            } else if constexpr (std::is_same_v<encoded_variable_t, four_byte_encoded_variable_t>) {
                epoch_time_ms_t timestamp_delta{};
                err = four_byte_encoding::decode_next_message(
                        ir_buf,
                        deserializer->m_log_event.m_log_message,
                        timestamp_delta
                );
                timestamp = deserializer->m_timestamp + timestamp_delta;
            } else {
                static_assert(cAlwaysFalse<encoded_variable_t>, "Invalid/unhandled encoding type");
            }
            if (IRErrorCode_Success != err) {
                return static_cast<int>(err);
            }
            deserializer->m_timestamp = timestamp;

            if (time_interval.m_upper <= deserializer->m_timestamp) {
                // TODO this is an extremely fragile hack until the CLP ffi ir
                // code is refactored and IRErrorCode includes things beyond
                // decoding.
                return static_cast<int>(IRErrorCode_Incomplete_IR + 1);
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
            size_t pos{0};
            if (ErrorCode_Success != ir_buf.try_get_pos(pos)) {
                return static_cast<int>(IRErrorCode_Decode_Error);
            }
            *ir_pos = pos;
            log_event->m_log_message.m_data = deserializer->m_log_event.m_log_message.data();
            log_event->m_log_message.m_size = deserializer->m_log_event.m_log_message.size();
            log_event->m_timestamp = deserializer->m_timestamp;
            *matching_query = matching_query_idx;
            return static_cast<int>(IRErrorCode_Success);
        }
    }
}  // namespace

extern "C" auto ir_deserializer_close(void* ir_deserializer) -> void {
    // NOLINTNEXTLINE(cppcoreguidelines-owning-memory)
    delete static_cast<Deserializer*>(ir_deserializer);
}

extern "C" auto ir_deserializer_deserialize_preamble(
        ByteSpan ir_view,
        size_t* ir_pos,
        int8_t* ir_encoding,
        int8_t* metadata_type,
        size_t* metadata_pos,
        uint16_t* metadata_size,
        void** ir_deserializer_ptr,
        void** timestamp_ptr
) -> int {
    BufferReader ir_buf{static_cast<char const*>(ir_view.m_data), ir_view.m_size};

    bool four_byte_encoding{};
    if (IRErrorCode const err{get_encoding_type(ir_buf, four_byte_encoding)};
        IRErrorCode_Success != err)
    {
        return static_cast<int>(err);
    }
    *ir_encoding = four_byte_encoding ? 1 : 0;

    if (IRErrorCode const err{
                decode_preamble(ir_buf, *metadata_type, *metadata_pos, *metadata_size)};
        IRErrorCode_Success != err)
    {
        return static_cast<int>(err);
    }

    size_t pos{0};
    if (ErrorCode_Success != ir_buf.try_get_pos(pos)) {
        return static_cast<int>(IRErrorCode_Decode_Error);
    }
    *ir_pos = pos;
    auto* deserializer{new Deserializer()};
    *ir_deserializer_ptr = deserializer;
    *timestamp_ptr = &deserializer->m_timestamp;
    return static_cast<int>(IRErrorCode_Success);
}

extern "C" auto ir_deserializer_deserialize_eight_byte_log_event(
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

extern "C" auto ir_deserializer_deserialize_four_byte_log_event(
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

extern "C" auto ir_deserializer_deserialize_eight_byte_wildcard_match(
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

extern "C" auto ir_deserializer_deserialize_four_byte_wildcard_match(
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
