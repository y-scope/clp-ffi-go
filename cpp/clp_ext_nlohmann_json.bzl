cc_library(
    name = "libjson",
    srcs = ["json/single_include/nlohmann/json.hpp"],
    hdrs = ["json/single_include/nlohmann/json.hpp"],
    includes = ["."],
    visibility = ["//visibility:public"],
)
