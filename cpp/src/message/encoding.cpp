#include <algorithm>
#include <memory>
#include <string>
#include <string_view>
#include <vector>

#include <LogEvent.hpp>
#include <clp/components/core/src/Defs.h>
#include <clp/components/core/src/ffi/encoding_methods.hpp>
#include <message/encoding.h>

// Note: dict_var_end_offsets is int32_t due to JNI putting limitations on
// encode_message.
struct EncodedMessage {
    std::string logtype;
    std::vector<encoded_variable_t> vars;
    std::vector<char> dict_vars;
    std::vector<int32_t> dict_var_end_offsets;
};

int decode_message(void* encoded_msg,
                   void** log_event_ptr,
                   char** log_event,
                   size_t* log_event_size) {
    EncodedMessage* em(reinterpret_cast<EncodedMessage*>(encoded_msg));
    auto event = std::make_unique<LogEvent>(em->logtype.size() * 2);
    event->msg = ffi::decode_message(
            em->logtype,
            em->vars.data(),
            em->vars.size(),
            std::string_view(reinterpret_cast<char const*>(em->dict_vars.data()),
                             em->dict_vars.size()),
            em->dict_var_end_offsets.data(),
            em->dict_var_end_offsets.size());

    *log_event = event->msg.data();
    *log_event_size = event->msg.size();
    *log_event_ptr = event.release();
    return 0;
}

void* encode_message(void* src_msg,
                     size_t src_size,
                     void** logtype,
                     void* logtype_size,
                     void** vars,
                     void* vars_size,
                     void** dict_vars,
                     void* dict_vars_size,
                     void** dict_var_end_offsets,
                     void* dict_var_end_offsets_size) {
    // We cannot use unique_ptr here as we want the Go code to hold any
    // references. Storing references in cpp (to avoid the unique_ptr falling
    // out of scope) means we need to synchronize the updates to that storage
    // as different go user threads could either encode a new message or free a
    // stored encoded message. We also cannot return/move a unique_ptr back up
    // to Go.
    EncodedMessage* em = new EncodedMessage();
    std::string_view msg(reinterpret_cast<char const*>(src_msg), src_size);

    std::vector<int32_t> dict_var_offsets;
    if (false == ffi::encode_message(msg, em->logtype, em->vars, dict_var_offsets)) {
        delete em;
        return nullptr;
    }

    // dict_var_offsets contains begin_pos followed by end_pos of each
    // dictionary variable in msg
    int32_t prev_end_off = 0;
    for (size_t i = 0; i < dict_var_offsets.size(); i += 2) {
        int32_t begin_pos = dict_var_offsets[i];
        int32_t end_pos = dict_var_offsets[i + 1];
        em->dict_vars.insert(em->dict_vars.begin() + prev_end_off,
                             msg.begin() + begin_pos,
                             msg.begin() + end_pos);
        prev_end_off = prev_end_off + (end_pos - begin_pos);
        em->dict_var_end_offsets.push_back(prev_end_off);
    }

    *logtype = em->logtype.data();
    *static_cast<std::size_t*>(logtype_size) = em->logtype.size();
    *vars = em->vars.data();
    *static_cast<std::size_t*>(vars_size) = em->vars.size();
    *dict_vars = em->dict_vars.data();
    *static_cast<std::size_t*>(dict_vars_size) = em->dict_vars.size();
    *dict_var_end_offsets = em->dict_var_end_offsets.data();
    *static_cast<std::size_t*>(dict_var_end_offsets_size) = em->dict_var_end_offsets.size();
    return em;
}

void delete_encoded_message(void* encoded_msg) { delete (EncodedMessage*)encoded_msg; }
