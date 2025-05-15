#ifndef CLP_FFI_GO_DEF_H
#define CLP_FFI_GO_DEF_H
// header must support C, making modernize checks inapplicable
// NOLINTBEGIN(modernize-deprecated-headers)
// NOLINTBEGIN(modernize-use-using)

#include <stdlib.h>

/**
 * A span of a byte array passed down through Cgo.
 */
typedef struct {
    void const* m_data;
    size_t m_size;
} ByteSpan;

// NOLINTEND(modernize-use-using)
// NOLINTEND(modernize-deprecated-headers)
#endif  // CLP_FFI_GO_DEF_H
