#!/bin/bash
# Obtains pre-built CLP FFI libraries.
# Tries in order: download from releases, build with Docker, build with task.
# Skips entirely if the library already exists locally.
#
# Usage: download-libs.sh [arch]
#   arch: target architecture (amd64, arm64). Defaults to current architecture.

set -eu

REPO="y-scope/clp-ffi-go"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
LIB_DIR="${ROOT_DIR}/pre-built/lib"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

# Determine target architecture
if [ $# -ge 1 ]; then
    ARCH="$1"
else
    # Default to current architecture
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64)  ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        arm64)   ARCH="arm64" ;;
        *)
            echo "Error: Unsupported architecture: $ARCH" >&2
            exit 1
            ;;
    esac
fi

# Validate architecture
case "$ARCH" in
    amd64|arm64) ;;
    *)
        echo "Error: Unsupported architecture: $ARCH (supported: amd64, arm64)" >&2
        exit 1
        ;;
esac

LIB_NAME="libclp-ffi-go_${OS}_${ARCH}.a"
LIB_PATH="${LIB_DIR}/${LIB_NAME}"

# Skip if library already exists
if [ -f "$LIB_PATH" ]; then
    echo "Library already exists: ${LIB_PATH}"
    exit 0
fi

echo "Library not found: ${LIB_NAME}"
mkdir -p "$LIB_DIR"

# Try to download from GitHub releases
download_release() {
    local tag="$1"
    local url="https://github.com/${REPO}/releases/download/${tag}/${LIB_NAME}"
    echo "Trying to download from: ${url}"
    curl -fsSL "$url" -o "$LIB_PATH" 2>/dev/null
}

downloaded=false

# Try 1: Download by git hash (if in a git repo)
if command -v git &>/dev/null && git rev-parse --git-dir &>/dev/null; then
    GIT_HASH=$(git rev-parse --short HEAD)
    echo "Attempting download for commit: ${GIT_HASH}"
    if download_release "$GIT_HASH"; then
        downloaded=true
    fi
fi

# Try 2: Build with Docker
if [ "$downloaded" = false ] && command -v docker &>/dev/null; then
    echo "Download failed. Attempting to build with Docker..."

    cd "$ROOT_DIR"
    DOCKER_ENV="-e HOST_UID=$(id -u) -e HOST_GID=$(id -g)"

    if [ "$ARCH" = "amd64" ]; then
        echo "Building for x86_64..."
        docker build -t clp-ffi-go-builder .
        docker run --rm $DOCKER_ENV -v "${ROOT_DIR}/pre-built:/output" clp-ffi-go-builder
    else
        echo "Building for ${ARCH} (requires QEMU)..."
        docker run --privileged --rm tonistiigi/binfmt --install "$ARCH" || true
        docker buildx build --platform "linux/${ARCH}" -t "clp-ffi-go-builder:${ARCH}" --load .
        docker run --rm $DOCKER_ENV -v "${ROOT_DIR}/pre-built:/output" "clp-ffi-go-builder:${ARCH}"
    fi
fi

# Try 3: Build with task (cmake)
if [ ! -f "$LIB_PATH" ] && command -v task &>/dev/null; then
    echo "Attempting to build with task..."

    cd "$ROOT_DIR"
    if task cpp:install-release; then
        echo "Built with task"
    else
        echo "Warning: task build failed" >&2
    fi
fi

# Final check
if [ -f "$LIB_PATH" ]; then
    echo "Success: ${LIB_PATH}"
else
    echo "Error: Failed to obtain library." >&2
    echo "To build from source, install one of:" >&2
    echo "  - Docker" >&2
    echo "  - task + cmake 3.23+ + C++20 compiler" >&2
    exit 1
fi
