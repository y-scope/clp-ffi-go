load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

_build_clp_ext_com_github_nlohmann_json = """
cc_library(
    name = "libjson",
    srcs = ["json/single_include/nlohmann/json.hpp"],
    hdrs = ["json/single_include/nlohmann/json.hpp"],
    includes = ["."],
    visibility = ["//visibility:public"],
)
"""

_build_com_github_y_scope_clp = """
cc_library(
    name = "libclp_ffi_core",
    srcs = [
        "clp/components/core/src/clp/BufferReader.cpp",
        "clp/components/core/src/clp/ffi/encoding_methods.cpp",
        "clp/components/core/src/clp/ffi/ir_stream/encoding_methods.cpp",
        "clp/components/core/src/clp/ffi/ir_stream/decoding_methods.cpp",
        "clp/components/core/src/clp/ffi/ir_stream/utils.cpp",
        "clp/components/core/src/clp/ir/parsing.cpp",
        "clp/components/core/src/clp/ReaderInterface.cpp",
        "clp/components/core/src/clp/string_utils/string_utils.cpp",
    ],
    hdrs = [
        "clp/components/core/src/clp/BufferReader.hpp",
        "clp/components/core/src/clp/Defs.h",
        "clp/components/core/src/clp/ErrorCode.hpp",
        "clp/components/core/src/clp/ReaderInterface.hpp",
        "clp/components/core/src/clp/ffi/encoding_methods.hpp",
        "clp/components/core/src/clp/ffi/encoding_methods.inc",
        "clp/components/core/src/clp/ffi/ir_stream/byteswap.hpp",
        "clp/components/core/src/clp/ffi/ir_stream/encoding_methods.hpp",
        "clp/components/core/src/clp/ffi/ir_stream/decoding_methods.hpp",
        "clp/components/core/src/clp/ffi/ir_stream/decoding_methods.inc",
        "clp/components/core/src/clp/ffi/ir_stream/protocol_constants.hpp",
        "clp/components/core/src/clp/ffi/ir_stream/utils.hpp",
        "clp/components/core/src/clp/ir/parsing.inc",
        "clp/components/core/src/clp/ir/parsing.hpp",
        "clp/components/core/src/clp/ir/types.hpp",
        "clp/components/core/src/clp/string_utils/string_utils.hpp",
        "clp/components/core/src/clp/TraceableException.hpp",
        "clp/components/core/src/clp/time_types.hpp",
        "clp/components/core/src/clp/type_utils.hpp",
    ],
    includes = [
        ".",
        "./clp/components/core/src/clp",
    ],
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
"""

def _clp_ext_com_github_nlohmann_json():
    commit = "fec56a1a16c6e1c1b1f4e116a20e79398282626c"
    commit_sha256 = "8cbda3504fd1624fbce641d3f6b884c76e5afead1fa6d6abfcbea4b734dc634b"
    http_archive(
        name = "clp_ext_com_github_nlohmann_json",
        sha256 = commit_sha256,
        urls = ["https://github.com/nlohmann/json/archive/{}.zip".format(commit)],
        strip_prefix = "json-{}".format(commit),
        add_prefix = "json",
        build_file_content = _build_clp_ext_com_github_nlohmann_json,
    )

def com_github_y_scope_clp():
    _clp_ext_com_github_nlohmann_json()

    commit = "3c1f0ad1c44b53d6c17fd7c1d578ec61616b5661"
    commit_sha256 = "1daaa432357ed470eb8a2b5e7c8b4064418fa0f5d89fd075c6f1b4aef1ac6319"
    http_archive(
        name = "com_github_y_scope_clp",
        sha256 = commit_sha256,
        urls = ["https://github.com/y-scope/clp/archive/{}.zip".format(commit)],
        strip_prefix = "clp-{}".format(commit),
        add_prefix = "clp",
        build_file_content = _build_com_github_y_scope_clp,
    )

def _clp_ffi_go_ext_deps_impl(_):
    com_github_y_scope_clp()

clp_ffi_go_ext_deps = module_extension(
    implementation = _clp_ffi_go_ext_deps_impl,
)
