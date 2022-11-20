#ifndef MESSAGE_ENCODING_H
#define MESSAGE_ENCODING_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>
#include <stdlib.h>

int decode_message(void* encoded_msg,
                   void** log_event_ptr,
                   char** log_event,
                   size_t* log_event_size);

void* encode_message(void* src_msg,
                     size_t src_size,
                     void** logtype,
                     void* logtype_size,
                     void** vars,
                     void* vars_size,
                     void** dict_vars,
                     void* dict_vars_size,
                     void** dict_var_end_offsets,
                     void* dict_var_end_offsets_size);

void delete_encoded_message(void* encoded_msg);

#ifdef __cplusplus
}
#endif

#endif // MESSAGE_ENCODING_H
