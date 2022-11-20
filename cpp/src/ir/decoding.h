#ifndef IR_DECODING_H
#define IR_DECODING_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>
#include <stdlib.h>

typedef int64_t epoch_time_ms_t;

int decode_preamble(void* buf_ptr,
                    size_t buf_size,
                    size_t* buf_offset,
                    int8_t* ir_encoding,
                    int8_t* metadata_type,
                    size_t* metadata_pos,
                    uint16_t* metadata_size);

int eight_byte_decode_next_log_event(void* buf_ptr,
                                     size_t buf_size,
                                     size_t* buf_offset,
                                     void** decoded_message_ptr,
                                     char** message,
                                     size_t* message_size,
                                     epoch_time_ms_t* timestamp);

int four_byte_decode_next_log_event(void* buf_ptr,
                                    size_t buf_size,
                                    size_t* buf_offset,
                                    void** message_ptr,
                                    char** message,
                                    size_t* message_size,
                                    epoch_time_ms_t* timestamp_delta);

#ifdef __cplusplus
}
#endif

#endif // IR_DECODING_H
