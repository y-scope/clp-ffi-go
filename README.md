# clp-ffi-go

[![CLP on Zulip](https://img.shields.io/badge/zulip-yscope--clp%20chat-1888FA?logo=zulip)](https://yscope-clp.zulipchat.com/)

This module provides Go packages to interface with [CLP's core features][clp-core] through CLP's FFI
(foreign function interface). For complete technical documentation, see the Go docs:
https://pkg.go.dev/github.com/y-scope/clp-ffi-go

For contributing and development instructions, see [CONTRIBUTING.md](CONTRIBUTING.md).

[clp-core]: https://github.com/y-scope/clp/tree/main/components/core

## Getting started

To add the module to your project run: `go get github.com/y-scope/clp-ffi-go`

### Reading log events

Read all log events from a zstd-compressed CLP IR stream:

```go
import (
  "os"

  "github.com/klauspost/compress/zstd"
  "github.com/y-scope/clp-ffi-go/ir"
)

file, _ := os.Open("log-file.clp.zst")
defer file.Close()
zstdReader, _ := zstd.NewReader(file)
defer zstdReader.Close()
irReader, _ := ir.NewReader(zstdReader)
defer irReader.Close()

for {
  event, err := irReader.ReadLogEvent()
  if ir.IrEndOfStream == err {
    break
  }
  if nil != err {
    log.Fatalf("ReadLogEvent failed: %v", err)
  }
  fmt.Println(event.UserKvPairs)
}
```

### Writing log events

Write log events to a CLP IR stream:

```go
import (
  "os"

  "github.com/y-scope/clp-ffi-go/ffi"
  "github.com/y-scope/clp-ffi-go/ir"
)

file, _ := os.Create("output.clp")
defer file.Close()
irWriter, _ := ir.NewWriter[ir.EightByteEncoding](file)
defer irWriter.Close()

event := ffi.LogEvent{
  AutoKvPairs: map[string]any{"timestamp": 1234567890},
  UserKvPairs: map[string]any{"level": "INFO", "message": "startup complete"},
}
irWriter.WriteLogEvent(event)
```

### Tracking field statistics

Register collectors on a Reader or Writer to automatically compute statistics as events flow
through. Collectors observe a named field from each log event and can be queried at any time.

```go
import (
  "fmt"
  "log"

  "github.com/y-scope/clp-ffi-go/ir"
)

irReader, _ := ir.NewReader(zstdReader)
defer irReader.Close()

// Track the "level" field (defaults to UserKvPairs).
irReader.Track("level", ir.NewUniqueCounts())

// Track a field from AutoKvPairs.
irReader.Track("host", ir.NewUniqueCounts(), ir.AutoField)

// Read events as normal — collectors update automatically.
for {
  _, err := irReader.ReadLogEvent()
  if ir.IrEndOfStream == err {
    break
  }
  if nil != err {
    log.Fatalf("ReadLogEvent failed: %v", err)
  }
}

// Query statistics at any point.
levelStats := irReader.Collector("level").(*ir.UniqueCounts)
fmt.Println("Unique levels:", levelStats.UniqueCount())
fmt.Println("Counts:", levelStats.ValueCounts())
// e.g. Unique levels: 3
//      Counts: map[ERROR:42 INFO:1503 WARN:87]
```

### Filtering with ReadToFunc

Use `ReadToFunc` to read until a predicate matches:

```go
// Find the first ERROR event.
event, err := irReader.ReadToFunc(func(e ffi.LogEvent) bool {
  return e.UserKvPairs["level"] == "ERROR"
})
```

## Bazel support

We provide a Bazel module and build files for each Go package. The following example shows how to
add the `ir` package as a dependency.

```bazel
# MODULE.bazel

bazel_dep(name = "com_github_y_scope_clp_ffi_go", version = "<version>")
_com_github_y_scope_clp_ffi_go_commit = "<commit hash>"
archive_override(
    module_name = "com_github_y_scope_clp_ffi_go",
    integrity = "sha256/512-<base 64 sha of commit>",
    urls = [
        "https://github.com/y-scope/clp-ffi-go/archive/{}.zip".format(
            _com_github_y_scope_clp_ffi_go_commit
        ),
    ],
    strip_prefix = "clp-ffi-go-{}".format(_com_github_y_scope_clp_ffi_go_commit),
)
# For local development:
# local_path_override(
#     module_name = "com_github_y_scope_clp_ffi_go",
#     path = "/home/user/clp-ffi-go",
# )
clp_ffi_go_ext_deps = use_extension(
    "@com_github_y_scope_clp_ffi_go//:bazel/deps.bzl", "clp_ffi_go_ext_deps"
)
use_repo(clp_ffi_go_ext_deps, "com_github_y_scope_clp")
use_repo(clp_ffi_go_ext_deps, "clp_ext_com_github_ned14_outcome")
```

```bazel
# BUILD.bazel

go_binary(
    name = "example",
    srcs = ["example.go"],
    visibility = ["//visibility:public"],
    deps = ["@com_github_y_scope_clp_ffi_go//ir"],
)
```
