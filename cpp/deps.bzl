load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def _clp_ext_com_github_nlohmann_json():
    commit = "fec56a1a16c6e1c1b1f4e116a20e79398282626c"
    commit_sha256 = "8cbda3504fd1624fbce641d3f6b884c76e5afead1fa6d6abfcbea4b734dc634b"
    http_archive(
        name = "clp_ext_com_github_nlohmann_json",
        sha256 = commit_sha256,
        urls = ["https://github.com/nlohmann/json/archive/{}.zip".format(commit)],
        strip_prefix = "json-{}".format(commit),
        add_prefix = "json",
        build_file_content = """
cc_library(
    name = "libjson",
    srcs = ["json/single_include/nlohmann/json.hpp"],
    hdrs = ["json/single_include/nlohmann/json.hpp"],
    includes = ["."],
    visibility = ["//visibility:public"],
)
        """,
    )

def com_github_y_scope_clp():
    _clp_ext_com_github_nlohmann_json()

    commit = "084efa35b7e9a63aecc5e327b97aea2a1cef83bc"
    commit_sha256 = "3aea613f00b8ca2e07803c5774a2faf8d7a315d983093eb4ce23a14a73414f72"
    http_archive(
        name = "com_github_y_scope_clp",
        sha256 = commit_sha256,
        urls = ["https://github.com/y-scope/clp/archive/{}.zip".format(commit)],
        strip_prefix = "clp-{}".format(commit),
        add_prefix = "clp",
        build_file_content = """
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
    ],
    deps = [
        "@clp_ext_com_github_nlohmann_json//:libjson",
    ],
    visibility = ["//visibility:public"],
)
        """,
    )

def _clp_ffi_go_ext_deps_impl(_):
    com_github_y_scope_clp()

clp_ffi_go_ext_deps = module_extension(
    implementation = _clp_ffi_go_ext_deps_impl,
)
