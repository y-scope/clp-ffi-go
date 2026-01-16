# Build environment for clp-ffi-go native library
# Produces portable libclp-ffi-go_linux_{amd64,arm64}.a
#
# Usage:
#   # Build for current architecture
#   docker build -t clp-ffi-go-builder .
#   docker run --rm -v $(pwd)/pre-built:/output clp-ffi-go-builder
#
#   # Build for specific architecture (uses QEMU)
#   docker buildx build --platform linux/arm64 -t clp-ffi-go-builder:arm64 .
#   docker run --rm -v $(pwd)/pre-built:/output clp-ffi-go-builder:arm64

FROM quay.io/pypa/manylinux_2_28_x86_64 AS builder-amd64
FROM quay.io/pypa/manylinux_2_28_aarch64 AS builder-arm64

# Select the right base image based on target platform
ARG TARGETARCH
FROM builder-${TARGETARCH} AS builder

# Install task
RUN curl -fsSL https://taskfile.dev/install.sh | sh -s -- -d -b /usr/local/bin

WORKDIR /src
COPY . .

# Build the library
RUN task cpp:build-release

# Copy to /output on run, fix ownership if HOST_UID is set
CMD task cpp:install-release INSTALL_PREFIX=/output && \
    if [ -n "$HOST_UID" ]; then chown -R "$HOST_UID:$HOST_GID" /output; fi && \
    echo "Built: $(ls /output/lib/)"
