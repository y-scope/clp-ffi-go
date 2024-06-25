cc_library(
    name = "libclp_ffi_core",
    srcs = [
        "clp/components/core/src/BufferReader.cpp",
        "clp/components/core/src/ReaderInterface.cpp",
        "clp/components/core/src/string_utils.cpp",
        "clp/components/core/src/ffi/encoding_methods.cpp",
        "clp/components/core/src/ffi/ir_stream/encoding_methods.cpp",
        "clp/components/core/src/ffi/ir_stream/decoding_methods.cpp",
    ],
    hdrs = [
        "clp/components/core/src/BufferReader.hpp",
        "clp/components/core/src/Defs.h",
        "clp/components/core/src/ErrorCode.hpp",
        "clp/components/core/src/ReaderInterface.hpp",
        "clp/components/core/src/string_utils.hpp",
        "clp/components/core/src/string_utils.inc",
        "clp/components/core/src/TraceableException.hpp",
        "clp/components/core/src/type_utils.hpp",
        "clp/components/core/src/ffi/encoding_methods.hpp",
        "clp/components/core/src/ffi/encoding_methods.inc",
        "clp/components/core/src/ffi/ir_stream/byteswap.hpp",
        "clp/components/core/src/ffi/ir_stream/encoding_methods.hpp",
        "clp/components/core/src/ffi/ir_stream/decoding_methods.hpp",
        "clp/components/core/src/ffi/ir_stream/decoding_methods.inc",
        "clp/components/core/src/ffi/ir_stream/protocol_constants.hpp",
    ],
    includes = ["."],
    copts = [
        "-std=c++20",
    ] + select({
        "@platforms//os:osx": [
            "-mmacosx-version-min=10.15",
        ],
        "//conditions:default": [],
    }),
    deps = [
        "@clp_ext_com_github_nlohmann_json//:libjson",
    ],
    visibility = ["//visibility:public"],
)

