clp-ffi-go
==========
.. image:: https://img.shields.io/badge/zulip-yscope--clp%20chat-1888FA?logo=zulip
   :alt: CLP on Zulip
   :target: https://yscope-clp.zulipchat.com/

This module provides Go packages to interface with `CLP's core features`__ through CLP's FFI
(foreign function interface). For complete technical documentation see the Go docs:
https://pkg.go.dev/github.com/y-scope/clp-ffi-go

__ https://github.com/y-scope/clp/tree/main/components/core

Getting started
---------------
To add the module to your project run: ``go get github.com/y-scope/clp-ffi-go``

Here's an example showing how to decode each log event containing "ERROR" from a CLP IR byte stream.

.. code:: golang

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

Building
--------
We use the ``go generate`` command to build the C++ interface to CLP's FFI code as well as stringify
``Enum`` style types.

1. Install requirements:

   a. A C++ compiler that supports C++20
   #. CMake 3.23 or higher
   #. The Stringer tool: https://pkg.go.dev/golang.org/x/tools/cmd/stringer

      - ``go install golang.org/x/tools/cmd/stringer@latest``

#. ``go generate ./...``

   - Run all generate directives (note the 3 dots after '/')

Bazel support
'''''''''''''
We provide a Bazel module and build files for each Go package in the repository.
Additionally, we provide a module extension for the FFI core component of CLP necessary to build the
native library.

The following is an example to pull in the ``ir`` Go package as a dependency through Bazel. For
development and testing it may be useful to use `git_override`_ or `local_path_override`_ to use
your own copy of the ffi-go repository.

.. _git_override: https://bazel.build/versions/6.0.0/rules/lib/globals#git_override

.. _local_path_override: https://bazel.build/versions/6.0.0/rules/lib/globals#local_path_override

.. code:: bazel

  # Add to MODULE.bazel

  bazel_dep(name = "com_github_y_scope_clp_ffi_go", version = "0.0.5-beta")
  clp_ffi_go_ext_deps = use_extension("@com_github_y_scope_clp_ffi_go//cpp:deps.bzl", "clp_ffi_go_ext_deps")
  use_repo(clp_ffi_go_ext_deps, "com_github_y_scope_clp")

.. code:: bazel

  # Add a ffi-go package as a dependency in a BUILD.bazel file

  go_binary(
      name = "example",
      srcs = ["example.go"],
      visibility = ["//visibility:public"],
      deps = ["@com_github_y_scope_clp_ffi_go//ir"],
  )

Why not build with cgo?
'''''''''''''''''''''''
The primary reason we choose to build with CMake rather than directly with cgo,
is to ease code maintenance by maximizing the reuse of CLP's code with no
modifications. If a platform you use is not supported by the pre-built
libraries, please open an issue and we can integrate it into our build process.

Testing
-------
To run all unit tests run: ``go_test_ir="/path/to/my-ir.clp.zst" go test ./...``

- Some of the ``ir`` package's tests currently require an existing CLP IR file
  compressed with zstd. This file's path is taken as an environment variable
  named ``go_test_ir``. It can be an absolute path or a path relative to the
  ``ir`` directory.

Linting
--------
1. Install golangci-lint:

.. code:: bash

    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
      sh -s -- -b $(go env GOPATH)/bin v1.59.0

2. Run with ``golangci-lint run``

Using an external C++ library
-----------------------------
Use the ``external`` build tag to link with different CLP FFI library instead
of the pre-built ones found in `lib`__. This tag only prevents the linking of
the pre-built libraries and does nothing else. It is up to the user to use
``CGO_LDFLAGS`` to point to their library. You may also need to update
``CGO_CFLAGS`` to update the header include path.

__ https://github.com/y-scope/clp-ffi-go/lib

For example, to run the tests using the ``external`` you can run:

.. code:: bash

  CGO_LDFLAGS="-L/path/to/external_libs -lclp_ffi_linux_amd64 -Wl,-rpath=/path/to/external_libs" \
    go_test_ir="/path/to/my-ir.clp.zst" \
    go test -tags external ./...
