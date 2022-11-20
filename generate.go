//go:generate cmake -S cpp -B cpp/build
//go:generate cmake --build cpp/build
//go:generate cmake -E env GOOS=${GOOS} GOARCH=${GOARCH} cmake --install cpp/build --prefix lib

package ffi
