#ifndef IR_ENCODING_H
#define IR_ENCODING_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>
#include <stdlib.h>

// TODO: replace with clp c-compatible header once it is created
typedef int64_t epoch_time_ms_t;

void* eight_byte_encode_preamble(void* ts_pattern_ptr,
                                 size_t ts_pattern_size,
                                 void* ts_pattern_syntax_ptr,
                                 size_t ts_pattern_syntax_size,
                                 void* time_zone_id_ptr,
                                 size_t time_zone_id_size,
                                 void** ir_buf_ptr,
                                 void* ir_buf_size);
void* four_byte_encode_preamble(void* ts_pattern_ptr,
                                size_t ts_pattern_size,
                                void* ts_pattern_syntax_ptr,
                                size_t ts_pattern_syntax_size,
                                void* time_zone_id_ptr,
                                size_t time_zone_id_size,
                                epoch_time_ms_t reference_ts,
                                void** ir_buf_ptr,
                                void* ir_buf_size);

int eight_byte_encode_message(void* irstream,
                              epoch_time_ms_t timestamp,
                              void* message_ptr,
                              size_t message_size,
                              void** ir_buf_ptr,
                              void* ir_buf_size);
int four_byte_encode_message(void* irstream,
                             epoch_time_ms_t timestamp_delta,
                             void* message_ptr,
                             size_t message_size,
                             void** ir_buf_ptr,
                             void* ir_buf_size);

void delete_ir_stream_state(void* irs);

#ifdef __cplusplus
}
#endif

#endif // IR_ENCODING_H
