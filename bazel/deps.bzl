load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

_build_clp_ext_com_github_ned14_outcome = """
cc_library(
    name = "outcome",
    hdrs = ["outcome/single-header/outcome.hpp"],
    includes = ["."],
    visibility = ["//visibility:public"],
)
"""

def clp_ext_com_github_ned14_outcome():
    ref = "2.2.10"
    ref_sha256 = "6505320e8d0e55913a9e3c6b33d03c61967429535fbb1fb8613c21527bb15110"
    http_archive(
        name = "clp_ext_com_github_ned14_outcome",
        sha256 = ref_sha256,
        urls = ["https://github.com/ned14/outcome/archive/v{}.tar.gz".format(ref)],
        strip_prefix = "outcome-{}".format(ref),
        add_prefix = "outcome",
        build_file_content = _build_clp_ext_com_github_ned14_outcome,
    )

_build_com_github_y_scope_clp = """
cc_library(
    name = "libclp_ffi_core",
    srcs = [
        "clp/components/core/src/clp/BufferReader.cpp",
        "clp/components/core/src/clp/ffi/encoding_methods.cpp",
        "clp/components/core/src/clp/ffi/ir_stream/encoding_methods.cpp",
        "clp/components/core/src/clp/ffi/ir_stream/decoding_methods.cpp",
        "clp/components/core/src/clp/ffi/ir_stream/Serializer.cpp",
        "clp/components/core/src/clp/ffi/ir_stream/utils.cpp",
        "clp/components/core/src/clp/ffi/SchemaTree.cpp",
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
        "clp/components/core/src/clp/ffi/KeyValuePairLogEvent.hpp",
        "clp/components/core/src/clp/ffi/ir_stream/byteswap.hpp",
        "clp/components/core/src/clp/ffi/ir_stream/Deserializer.hpp",
        "clp/components/core/src/clp/ffi/ir_stream/encoding_methods.hpp",
        "clp/components/core/src/clp/ffi/ir_stream/decoding_methods.hpp",
        "clp/components/core/src/clp/ffi/ir_stream/decoding_methods.inc",
        "clp/components/core/src/clp/ffi/ir_stream/ir_unit_deserialization_methods.hpp",
        "clp/components/core/src/clp/ffi/ir_stream/IrUnitHandlerInterface.hpp",
        "clp/components/core/src/clp/ffi/ir_stream/IrUnitType.hpp",
        "clp/components/core/src/clp/ffi/ir_stream/protocol_constants.hpp",
        "clp/components/core/src/clp/ffi/ir_stream/Serializer.hpp",
        "clp/components/core/src/clp/ffi/ir_stream/utils.hpp",
        "clp/components/core/src/clp/ffi/SchemaTree.hpp",
        "clp/components/core/src/clp/ffi/Value.hpp",
        "clp/components/core/src/clp/ir/EncodedTextAst.hpp",
        "clp/components/core/src/clp/ir/parsing.inc",
        "clp/components/core/src/clp/ir/parsing.hpp",
        "clp/components/core/src/clp/ir/types.hpp",
        "clp/components/core/src/clp/string_utils/string_utils.hpp",
        "clp/components/core/src/clp/TraceableException.hpp",
        "clp/components/core/src/clp/time_types.hpp",
        "clp/components/core/src/clp/type_utils.hpp",
    ],
    includes = [
        "clp/components/core/src",
        "clp/components/core/src/clp",
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
        "@clp_ext_com_github_ned14_outcome//:outcome",
        "@nlohmann_json//:singleheader-json",
        "@msgpack-c//:msgpack",
    ],
    visibility = ["//visibility:public"],
)
"""

def com_github_y_scope_clp():
    ref = "e21672b906641c4724a25ea74f13857afdebe0e8"
    ref_sha256 = "b7ab19af62fb0601d858047452e2f330489070caccd4aaf1e09709f6ca6324ab"
    http_archive(
        name = "com_github_y_scope_clp",
        sha256 = ref_sha256,
        urls = ["https://github.com/y-scope/clp/archive/{}.tar.gz".format(ref)],
        strip_prefix = "clp-{}".format(ref),
        add_prefix = "clp",
        build_file_content = _build_com_github_y_scope_clp,
    )

def _clp_ffi_go_ext_deps_impl(_):
    clp_ext_com_github_ned14_outcome()
    com_github_y_scope_clp()

clp_ffi_go_ext_deps = module_extension(
    implementation = _clp_ffi_go_ext_deps_impl,
)
