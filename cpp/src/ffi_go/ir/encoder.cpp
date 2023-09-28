#include "encoder.h"

#include <algorithm>
#include <memory>
#include <string>
#include <string_view>
#include <type_traits>
#include <vector>

#include <clp/components/core/src/ffi/encoding_methods.hpp>
#include <clp/components/core/src/ffi/ir_stream/decoding_methods.hpp>

#include <ffi_go/defs.h>
#include <ffi_go/ir/LogTypes.hpp>
#include <ffi_go/LogTypes.hpp>

namespace ffi_go::ir {
using namespace ffi::ir_stream;

namespace {
    /**
     * Generic helper for ir_encoder_encode_*_log_message
     */
    template <class encoded_var_view_t>
    auto encode_log_message(
            StringView log_message,
            void* ir_encoder,
            StringView* logtype,
            encoded_var_view_t* vars,
            StringView* dict_vars,
            Int32tSpan* dict_var_end_offsets
    ) -> int {
        typedef typename std::conditional<
                std::is_same_v<Int64tSpan, encoded_var_view_t>,
                ffi::eight_byte_encoded_variable_t,
                ffi::four_byte_encoded_variable_t>::type encoded_var_t;
        Encoder<encoded_var_t>* encoder{static_cast<Encoder<encoded_var_t>*>(ir_encoder)};
        LogMessage<encoded_var_t>& ir_log_msg = encoder->m_log_message;
        ir_log_msg.reserve(log_message.m_size);

        std::string_view const log_msg_view{log_message.m_data, log_message.m_size};
        std::vector<int32_t> dict_var_offsets;
        if (false
            == ffi::encode_message<encoded_var_t>(
                    log_msg_view,
                    ir_log_msg.m_logtype,
                    ir_log_msg.m_vars,
                    dict_var_offsets
            ))
        {
            return static_cast<int>(IRErrorCode_Corrupted_IR);
        }

        // dict_var_offsets contains begin_pos followed by end_pos of each
        // dictionary variable in the message
        int32_t prev_end_off = 0;
        for (size_t i = 0; i < dict_var_offsets.size(); i += 2) {
            int32_t const begin_pos = dict_var_offsets[i];
            int32_t const end_pos = dict_var_offsets[i + 1];
            ir_log_msg.m_dict_vars.insert(
                    ir_log_msg.m_dict_vars.begin() + prev_end_off,
                    log_msg_view.begin() + begin_pos,
                    log_msg_view.begin() + end_pos
            );
            prev_end_off = prev_end_off + (end_pos - begin_pos);
            ir_log_msg.m_dict_var_end_offsets.push_back(prev_end_off);
        }

        logtype->m_data = ir_log_msg.m_logtype.data();
        logtype->m_size = ir_log_msg.m_logtype.size();
        vars->m_data = ir_log_msg.m_vars.data();
        vars->m_size = ir_log_msg.m_vars.size();
        dict_vars->m_data = ir_log_msg.m_dict_vars.data();
        dict_vars->m_size = ir_log_msg.m_dict_vars.size();
        dict_var_end_offsets->m_data = ir_log_msg.m_dict_var_end_offsets.data();
        dict_var_end_offsets->m_size = ir_log_msg.m_dict_var_end_offsets.size();
        return static_cast<int>(IRErrorCode_Success);
    }
}  // namespace

extern "C" auto ir_encoder_eight_byte_new() -> void* {
    return new Encoder<ffi::eight_byte_encoded_variable_t>{};
}

extern "C" auto ir_encoder_four_byte_new() -> void* {
    return new Encoder<ffi::four_byte_encoded_variable_t>{};
}

extern "C" auto ir_encoder_eight_byte_close(void* ir_encoder) -> void {
    // NOLINTNEXTLINE(cppcoreguidelines-owning-memory)
    delete static_cast<Encoder<ffi::eight_byte_encoded_variable_t>*>(ir_encoder);
}

extern "C" auto ir_encoder_four_byte_close(void* ir_encoder) -> void {
    // NOLINTNEXTLINE(cppcoreguidelines-owning-memory)
    delete static_cast<Encoder<ffi::four_byte_encoded_variable_t>*>(ir_encoder);
}

extern "C" auto ir_encoder_encode_eight_byte_log_message(
        StringView log_message,
        void* ir_encoder,
        StringView* logtype,
        Int64tSpan* vars_ptr,
        StringView* dict_vars,
        Int32tSpan* dict_var_end_offsets
) -> int {
    return encode_log_message(
            log_message,
            ir_encoder,
            logtype,
            vars_ptr,
            dict_vars,
            dict_var_end_offsets
    );
}

extern "C" auto ir_encoder_encode_four_byte_log_message(
        StringView log_message,
        void* ir_encoder,
        StringView* logtype,
        Int32tSpan* vars,
        StringView* dict_vars,
        Int32tSpan* dict_var_end_offsets
) -> int {
    return encode_log_message(
            log_message,
            ir_encoder,
            logtype,
            vars,
            dict_vars,
            dict_var_end_offsets
    );
}
}  // namespace ffi_go::ir
