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
        build_file = "//:clp_ext_nlohmann_json.bzl",
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
        build_file = "//:clp_ffi_core.bzl",
    )

def _clp_ffi_go_ext_deps_impl(_):
    com_github_y_scope_clp()

clp_ffi_go_ext_deps = module_extension(
    implementation = _clp_ffi_go_ext_deps_impl,
)
