package rapidyenc

/*
#cgo noescape rapidyenc_version
#cgo nocallback rapidyenc_version
#cgo noescape rapidyenc_decode_init
#cgo nocallback rapidyenc_decode_init
#cgo noescape rapidyenc_decode_kernel
#cgo nocallback rapidyenc_decode_kernel
#cgo noescape rapidyenc_decode_incremental_go
#cgo nocallback rapidyenc_decode_incremental_go
#cgo noescape rapidyenc_encode_init
#cgo nocallback rapidyenc_encode_init
#cgo noescape rapidyenc_encode_kernel
#cgo nocallback rapidyenc_encode_kernel
#cgo noescape rapidyenc_encode
#cgo nocallback rapidyenc_encode
#cgo noescape rapidyenc_encode_ex
#cgo nocallback rapidyenc_encode_ex
#cgo darwin,!ios LDFLAGS: ${SRCDIR}/librapidyenc_darwin.a -lstdc++
// Device and simulator slices need different archives: scripts/build_ios.sh
// builds them and passes the one for each slice in CGO_LDFLAGS.
#cgo ios LDFLAGS: -lc++
#cgo windows,amd64 LDFLAGS: ${SRCDIR}/librapidyenc_windows_amd64.a -lstdc++
#cgo linux,!android,amd64 LDFLAGS: ${SRCDIR}/librapidyenc_linux_amd64.a -lstdc++
#cgo linux,!android,arm64 LDFLAGS: ${SRCDIR}/librapidyenc_linux_arm64.a -lstdc++
#cgo android,arm64 LDFLAGS: ${SRCDIR}/librapidyenc_android_arm64.a -lc++_static -lc++abi
#cgo android,arm LDFLAGS: ${SRCDIR}/librapidyenc_android_arm.a -lc++_static -lc++abi
#cgo android,amd64 LDFLAGS: ${SRCDIR}/librapidyenc_android_amd64.a -lc++_static -lc++abi
#cgo android,386 LDFLAGS: ${SRCDIR}/librapidyenc_android_386.a -lc++_static -lc++abi
#include "rapidyenc.h"
*/
import "C"
