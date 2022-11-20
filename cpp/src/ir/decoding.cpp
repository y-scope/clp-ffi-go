#include "encoding.h"
#include <cstdint>
#include <memory>
#include <type_traits>

#include <LogEvent.hpp>
#include <clp/components/core/src/ffi/ir_stream/decoding_methods.hpp>
#include <ir/decoding.h>

using namespace ffi::ir_stream;

int decode_preamble(void* buf_ptr,
                    size_t buf_size,
                    size_t* buf_offset,
                    int8_t* ir_encoding,
                    int8_t* metadata_type,
                    size_t* metadata_pos,
                    uint16_t* metadata_size) {
    IrBuffer ir_buf{reinterpret_cast<int8_t*>(buf_ptr), buf_size};
    ir_buf.set_cursor_pos(*buf_offset);

    bool four_byte_encoding;
    if (IRErrorCode err{get_encoding_type(ir_buf, four_byte_encoding)};
        IRErrorCode_Success != err) {
        return static_cast<int>(err);
    }
    *ir_encoding = four_byte_encoding ? 1 : 0;

    if (IRErrorCode err{decode_preamble(ir_buf, *metadata_type, *metadata_pos, *metadata_size)};
        IRErrorCode_Success != err) {
        return static_cast<int>(err);
    }

    *buf_offset = ir_buf.get_cursor_pos();
    return static_cast<int>(IRErrorCode_Success);
}

int decode_next_log_event(IRErrorCode (*decode_fp)(IrBuffer&, std::string&, epoch_time_ms_t&),
                          void* buf_ptr,
                          size_t buf_size,
                          size_t* buf_offset,
                          void** log_event_ptr,
                          char** log_event,
                          size_t* log_event_size,
                          epoch_time_ms_t* timestamp) {
    IrBuffer ir_buf{reinterpret_cast<int8_t*>(buf_ptr), buf_size};
    ir_buf.set_cursor_pos(*buf_offset);
    auto event = std::make_unique<LogEvent>(buf_size);
    if (IRErrorCode err{decode_fp(ir_buf, event->msg, *timestamp)}; IRErrorCode_Success != err) {
        return static_cast<int>(err);
    }
    *buf_offset = ir_buf.get_cursor_pos();

    *log_event = event->msg.data();
    *log_event_size = event->msg.size();
    *log_event_ptr = event.release();
    return static_cast<int>(IRErrorCode_Success);
}

int eight_byte_decode_next_log_event(void* buf_ptr,
                                     size_t buf_size,
                                     size_t* buf_offset,
                                     void** log_event_ptr,
                                     char** log_event,
                                     size_t* log_event_size,
                                     epoch_time_ms_t* timestamp) {
    return decode_next_log_event(eight_byte_encoding::decode_next_message,
                                 buf_ptr,
                                 buf_size,
                                 buf_offset,
                                 log_event_ptr,
                                 log_event,
                                 log_event_size,
                                 timestamp);
}

int four_byte_decode_next_log_event(void* buf_ptr,
                                    size_t buf_size,
                                    size_t* buf_offset,
                                    void** log_event_ptr,
                                    char** log_event,
                                    size_t* log_event_size,
                                    epoch_time_ms_t* timestamp_delta) {
    return decode_next_log_event(four_byte_encoding::decode_next_message,
                                 buf_ptr,
                                 buf_size,
                                 buf_offset,
                                 log_event_ptr,
                                 log_event,
                                 log_event_size,
                                 timestamp_delta);
}
