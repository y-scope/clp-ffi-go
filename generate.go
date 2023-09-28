//go:generate cmake -E env GOOS=${GOOS} GOARCH=${GOARCH} cmake -S cpp -B cpp/build
//go:generate cmake --build cpp/build -j
//go:generate cmake --install cpp/build --prefix .

package clp_ffi_go
