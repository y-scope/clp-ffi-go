clp-ffi-go
==========
.. image:: https://img.shields.io/badge/zulip-yscope--clp%20chat-1888FA?logo=zulip
   :alt: CLP on Zulip
   :target: https://yscope-clp.zulipchat.com/

This module provides Go packages to interface with `CLP's core features`__
through CLP's FFI (foreign function interface). For complete technical
documentation see the Go docs: https://pkg.go.dev/github.com/y-scope/clp-ffi-go

__ https://github.com/y-scope/clp/tree/main/components/core

Getting started
---------------
To add the module to your project run: ``go get github.com/y-scope/clp-ffi-go``

Here's an example showing how to decode each log event containing "ERROR" from
a CLP IR byte stream.

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
We use the ``go generate`` command to build the C++ interface to CLP's FFI code
as well as stringify ``Enum`` style types.

1. Install requirements:

   a. A C++ compiler that supports C++17
   #. CMake 3.11 or higher
   #. The Stringer tool: https://pkg.go.dev/golang.org/x/tools/cmd/stringer

      - ``go install golang.org/x/tools/cmd/stringer@latest``

#. ``go generate ./...``

   - Run all generate directives (note the 3 dots after '/')

Bazel support
'''''''''''''
We provide Bazel build files for each Go package in the repository, enabling
you to add any package to your `build dependency list`__ with no extra
arguments or modifications.

__ https://github.com/bazelbuild/rules_go/blob/master/docs/go/core/rules.md#go_library-deps

Testing
-------
To run all unit tests run: ``go_test_ir="/path/to/my-ir.clp.zst" go test ./...``

- Some of the ``ir`` package's tests currently require an existing CLP IR file
  compressed with zstd. This file's path is taken as an environment variable
  named ``go_test_ir``. It can be an absolute path or a path relative to the
  ``ir`` directory.

Why not build with cgo?
'''''''''''''''''''''''
The primary reason we choose to build with CMake rather than directly with cgo,
is to ease code maintenance by maximizing the reuse of CLP's code with no
modifications. If a platform you use is not supported by the pre-built
libraries, please open an issue and we can integrate it into our build process.

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
