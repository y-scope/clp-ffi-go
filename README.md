# clp-ffi-go

[![CLP on Zulip](https://img.shields.io/badge/zulip-yscope--clp%20chat-1888FA?logo=zulip)](https://yscope-clp.zulipchat.com/)

This module provides Go packages to interface with [CLP's core features][clp-core] through CLP's FFI
(foreign function interface). For complete technical documentation, see the Go docs:
https://pkg.go.dev/github.com/y-scope/clp-ffi-go

[clp-core]: https://github.com/y-scope/clp/tree/main/components/core

## Getting started

To add the module to your project run: `go get github.com/y-scope/clp-ffi-go`

Here's an example showing how to decode each log event containing "ERROR" from a CLP IR byte stream.

```go
import (
  "fmt"
  "time"

  "github.com/klauspost/compress/zstd"
  "github.com/y-scope/clp-ffi-go/ffi"
  "github.com/y-scope/clp-ffi-go/ir"
)

file, _ := os.Open("log-file.clp.zst")
defer file.Close()
zstdReader, _ := zstd.NewReader(file)
defer zstdReader.Close()
irReader, _ := ir.NewReader(zstdReader)
defer irReader.Close()

var err error
for {
  var log *ffi.LogEventView
  // To read every log event replace ReadToContains with Read()
  log, err = irReader.ReadToContains("ERROR")
  if nil != err {
    break
  }
  fmt.Printf("%v %v", time.UnixMilli(int64(log.Timestamp)), log.LogMessageView)
}
if ir.EndOfIr != err {
  fmt.Printf("Reader.Read failed: %v", err)
}
```

## Building from source

Run `go generate ./...` to build the native libraries and generate Go code. The generate step will
automatically try to:

1. Download pre-built libraries from GitHub releases (matching your commit)
2. Build with Docker if no release is available
3. Build with task/cmake as a last resort

**Requirements** (only needed if Docker is unavailable):
- C++20 compiler
- CMake 3.23+
- [Task](https://taskfile.dev/installation/)
- [Stringer](https://pkg.go.dev/golang.org/x/tools/cmd/stringer): `go install golang.org/x/tools/cmd/stringer@latest`

## Development

### Testing

```bash
go_test_ir="/path/to/my-ir.clp.zst" go test ./...
```

Some tests in the `ir` package require an existing CLP IR file compressed with zstd. The path is
provided via the `go_test_ir` environment variable (absolute or relative to the `ir` directory).

### Linting

```bash
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
  sh -s -- -b $(go env GOPATH)/bin v1.59.0

# Run linter
golangci-lint run
```

### Building native libraries

The download script builds for your current architecture by default. Go's cgo build constraints
ensure only the matching library is linked.

```bash
# Build for current architecture
./scripts/download-libs.sh

# Build for specific architecture
./scripts/download-libs.sh amd64
./scripts/download-libs.sh arm64
```

To cross-compile Go for a different architecture:

```bash
# On arm64, build for x86_64
./scripts/download-libs.sh amd64
GOARCH=amd64 go build ./...

# On x86_64, build for arm64
./scripts/download-libs.sh arm64
GOARCH=arm64 go build ./...
```

For manual Docker builds:

```bash
# Build for x86_64
docker build -t clp-ffi-go-builder .
docker run --rm -v $(pwd)/pre-built:/output clp-ffi-go-builder

# Build for arm64 (uses QEMU emulation on x86_64 hosts)
docker run --privileged --rm tonistiigi/binfmt --install arm64
docker buildx build --platform linux/arm64 -t clp-ffi-go-builder:arm64 --load .
docker run --rm -v $(pwd)/pre-built:/output clp-ffi-go-builder:arm64
```

## Advanced

### Bazel support

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
    "@com_github_y_scope_clp_ffi_go//cpp:deps.bzl", "clp_ffi_go_ext_deps"
)
use_repo(clp_ffi_go_ext_deps, "com_github_y_scope_clp")
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

### Using an external C++ library

Use the `external` build tag to link with a different CLP FFI library. This disables the default
library linking; you must set `CGO_LDFLAGS` (and optionally `CGO_CFLAGS`) to point to your library.

```bash
CGO_LDFLAGS="-L/path/to/external_libs -lclp_ffi_linux_amd64 -Wl,-rpath=/path/to/external_libs" \
  go_test_ir="/path/to/my-ir.clp.zst" \
  go test -tags external ./...
```

### Why CMake instead of cgo?

We build with CMake rather than directly with cgo to maximize reuse of CLP's existing code without
modifications. If your platform is not supported by the pre-built libraries, please open an issue.
