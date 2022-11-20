#ifndef LOG_EVENT_HPP
#define LOG_EVENT_HPP

#include <string>

struct LogEvent {
    LogEvent() : msg{} {}
    LogEvent(size_t cap) { msg.reserve(cap); }

    std::string msg;
};

#endif // LOG_EVENT_HPP
