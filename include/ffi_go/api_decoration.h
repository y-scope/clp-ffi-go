#ifndef FFI_GO_API_DECORATION_H
#define FFI_GO_API_DECORATION_H

/**
 * If the file is compiled with a C++ compiler, `extern "C"` must be defined to
 * ensure C linkage.
 */
#ifdef __cplusplus
    #define CLP_FFI_GO_EXTERN_C extern "C"
#else
    #define CLP_FFI_GO_EXTERN_C
#endif

/**
 * `CLP_FFI_GO_METHOD` should be added at the beginning of a function's
 * declaration/implementation to decorate any APIs that are exposed to the
 * Golang layer.
 */
#define CLP_FFI_GO_METHOD CLP_FFI_GO_EXTERN_C

#endif
