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
a CLP IR stream.

.. code:: golang

  import (
    "fmt"
    "time"

    "github.com/klauspost/compress/zstd"
    "github.com/y-scope/clp-ffi-go/ir"
  )

  file, _ := os.Open("log-file.clp.zst")
  zstdReader, _ := zstd.NewReader(file)

  irReader, _ := ir.ReadPreamble(zstdReader, 4096)
  for {
    // To read every log event replace ReadToContains with
    // ReadNextLogEvent(zstdReader)
    log, err := irReader.ReadToContains(zstdReader, []byte("ERROR"))
    if io.EOF == err {
      break
    }
    fmt.Printf("%v %v", time.UnixMilli(int64(log.Timestamp)), string(log.Msg))
  }

Building
--------
We use the ``go generate`` command to build the C++ interface to CLP's FFI code
as well as stringify ``Enum`` style types.

1. Install requirements:

   a. A C++ compiler that supports C++17
   #. CMake 3.5.1 or higher
   #. The Stringer tool: https://pkg.go.dev/golang.org/x/tools/cmd/stringer

      - ``go install golang.org/x/tools/cmd/stringer@latest``

#. ``git submodule update --init --recursive``

   - Pull all submodules in preparation for building

#. ``go generate ./...``

   - Run all generate directives (note the 3 dots after '/')

Testing
-------
To run all unit tests run: ``go test ./... -args $(readlink -f clp-ir-stream.clp.zst)``

- The ``ir`` package's tests currently requries an existing CLP IR file
  compressed with zstd. This file's path is taken as the only argument to the
  test and is supplied after ``-args``. It can be an absolute path or a path
  relative to the ``ir`` directory.

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

  CGO_LDFLAGS="-L./lib -lclp_ffi_linux_amd64 -lstdc++" \
  go test -tags external,test ./... -args $(readlink -f clp-ir-stream.clp.zst)
