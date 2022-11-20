#include <log_event.h>
#include <LogEvent.hpp>

void delete_log_event(void* log_event) {
    delete reinterpret_cast<LogEvent*>(log_event);
}
