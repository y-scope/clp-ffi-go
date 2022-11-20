#include <string>
#include <string_view>
#include <vector>

#include <clp/components/core/src/ffi/encoding_methods.hpp>
#include <clp/components/core/src/ffi/ir_stream/encoding_methods.hpp>
#include <ir/encoding.h>

struct IrStreamState {
    std::string logtype;
    std::vector<int8_t> ir_buf;
};

void* eight_byte_encode_preamble(void* ts_pattern_ptr,
                                 size_t ts_pattern_size,
                                 void* ts_pattern_syntax_ptr,
                                 size_t ts_pattern_syntax_size,
                                 void* time_zone_id_ptr,
                                 size_t time_zone_id_size,
                                 void** ir_buf_ptr,
                                 void* ir_buf_size) {
    std::string_view ts_pattern(reinterpret_cast<char const*>(ts_pattern_ptr), ts_pattern_size);
    std::string_view ts_pattern_syntax(reinterpret_cast<char const*>(ts_pattern_syntax_ptr),
                                       ts_pattern_syntax_size);
    std::string_view time_zone_id(reinterpret_cast<char const*>(time_zone_id_ptr),
                                  time_zone_id_size);

    IrStreamState* irs = new IrStreamState();
    if (false == ffi::ir_stream::eight_byte_encoding::encode_preamble(
                         ts_pattern, ts_pattern_syntax, time_zone_id, irs->ir_buf)) {
        delete irs;
        return nullptr;
    }

    *ir_buf_ptr = irs->ir_buf.data();
    *static_cast<std::size_t*>(ir_buf_size) = irs->ir_buf.size();
    return irs;
}

void* four_byte_encode_preamble(void* ts_pattern_ptr,
                                size_t ts_pattern_size,
                                void* ts_pattern_syntax_ptr,
                                size_t ts_pattern_syntax_size,
                                void* time_zone_id_ptr,
                                size_t time_zone_id_size,
                                ffi::epoch_time_ms_t reference_ts,
                                void** ir_buf_ptr,
                                void* ir_buf_size) {
    std::string_view ts_pattern(reinterpret_cast<char const*>(ts_pattern_ptr), ts_pattern_size);
    std::string_view ts_pattern_syntax(reinterpret_cast<char const*>(ts_pattern_syntax_ptr),
                                       ts_pattern_syntax_size);
    std::string_view time_zone_id(reinterpret_cast<char const*>(time_zone_id_ptr),
                                  time_zone_id_size);

    IrStreamState* irs = new IrStreamState();
    if (false == ffi::ir_stream::four_byte_encoding::encode_preamble(
                         ts_pattern, ts_pattern_syntax, time_zone_id, reference_ts, irs->ir_buf)) {
        delete irs;
        return nullptr;
    }

    *ir_buf_ptr = irs->ir_buf.data();
    *static_cast<std::size_t*>(ir_buf_size) = irs->ir_buf.size();
    return irs;
}

int encode_message(
        bool (*em_fp)(ffi::epoch_time_ms_t, std::string_view, std::string&, std::vector<int8_t>&),
        void* irstream,
        ffi::epoch_time_ms_t timestamp_or_delta,
        void* message_ptr,
        size_t message_size,
        void** ir_buf_ptr,
        void* ir_buf_size) {
    IrStreamState* irs(reinterpret_cast<IrStreamState*>(irstream));
    std::string_view message(reinterpret_cast<char const*>(message_ptr), message_size);
    irs->ir_buf.clear();
    if (false == em_fp(timestamp_or_delta, message, irs->logtype, irs->ir_buf)) {
        return -1;
    }
    *ir_buf_ptr = irs->ir_buf.data();
    *static_cast<std::size_t*>(ir_buf_size) = irs->ir_buf.size();
    return 0;
}

int eight_byte_encode_message(void* irstream,
                              ffi::epoch_time_ms_t timestamp,
                              void* message_ptr,
                              size_t message_size,
                              void** ir_buf_ptr,
                              void* ir_buf_size) {
    return encode_message(ffi::ir_stream::eight_byte_encoding::encode_message,
                          irstream,
                          timestamp,
                          message_ptr,
                          message_size,
                          ir_buf_ptr,
                          ir_buf_size);
}

int four_byte_encode_message(void* irstream,
                             ffi::epoch_time_ms_t timestamp_delta,
                             void* message_ptr,
                             size_t message_size,
                             void** ir_buf_ptr,
                             void* ir_buf_size) {
    return encode_message(ffi::ir_stream::four_byte_encoding::encode_message,
                          irstream,
                          timestamp_delta,
                          message_ptr,
                          message_size,
                          ir_buf_ptr,
                          ir_buf_size);
}

void delete_ir_stream_state(void* irs) { delete (IrStreamState*)irs; }
