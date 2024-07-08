#ifndef FFI_GO_DEF_H
#define FFI_GO_DEF_H
// header must support C, making modernize checks inapplicable
// NOLINTBEGIN(modernize-deprecated-headers)
// NOLINTBEGIN(modernize-use-using)

#ifndef __cplusplus
    #include <stdbool.h>
#endif

#include <stdint.h>
#include <stdlib.h>

// TODO: replace with clp c-compatible header once it exists
typedef int64_t epoch_time_ms_t;

/**
 * A span of a bool array passed down through Cgo.
 */
typedef struct {
    bool* m_data;
    size_t m_size;
} BoolSpan;

/**
 * A span of a byte array passed down through Cgo.
 */
typedef struct {
    void* m_data;
    size_t m_size;
} ByteSpan;

/**
 * A span of a Go int32 array passed down through Cgo.
 */
typedef struct {
    int32_t* m_data;
    size_t m_size;
} Int32tSpan;

/**
 * A span of a Go int64 array passed down through Cgo.
 */
typedef struct {
    int64_t* m_data;
    size_t m_size;
} Int64tSpan;

/**
 * A span of a Go int/C.size_t array passed down through Cgo.
 */
typedef struct {
    size_t* m_data;
    size_t m_size;
} SizetSpan;

/**
 * A view of a Go string passed down through Cgo.
 */
typedef struct {
    char const* m_data;
    size_t m_size;
} StringView;

/**
 * A view of a Go ffi.LogEvent passed down through Cgo.
 */
typedef struct {
    StringView m_log_message;
    epoch_time_ms_t m_timestamp;
    epoch_time_ms_t m_utc_offset;
} LogEventView;

// NOLINTEND(modernize-use-using)
// NOLINTEND(modernize-deprecated-headers)
#endif  // FFI_GO_DEF_H
