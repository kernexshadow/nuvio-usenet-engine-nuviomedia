#!/usr/bin/env bash
# Builds build/ios/NuvioUsenet.xcframework: the engine as a static library for
# iOS devices and simulators, for apps that cannot start the Android-style
# sidecar process. Runs on macOS with Xcode, Go and CMake.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="$ROOT/build/ios"
MIN_IOS="${MIN_IOS:-16.1}"
GO="${GO:-go}"

rm -rf "$OUT"
mkdir -p "$OUT/headers"
cp "$ROOT/cmd/nuvio-usenet-ios/nuvio_usenet.h" "$OUT/headers/"

# build_slice NAME SDK GOARCH CMAKE_ARCH CLANG_TARGET
build_slice() {
  local name="$1" sdk="$2" goarch="$3" arch="$4" target="$5"
  local work="$OUT/$name" sdk_path cc cxx
  sdk_path="$(xcrun --sdk "$sdk" --show-sdk-path)"
  cc="$(xcrun --sdk "$sdk" --find clang)"
  cxx="$(xcrun --sdk "$sdk" --find clang++)"
  mkdir -p "$work"

  cmake -S "$ROOT/third_party/rapidyenc-native" -B "$work/rapidyenc" \
    -DCMAKE_SYSTEM_NAME=iOS -DCMAKE_OSX_SYSROOT="$sdk" -DCMAKE_OSX_ARCHITECTURES="$arch" \
    -DCMAKE_OSX_DEPLOYMENT_TARGET="$MIN_IOS" -DTARGET_ARCH="$arch" \
    -DCMAKE_TRY_COMPILE_TARGET_TYPE=STATIC_LIBRARY -DCMAKE_POLICY_VERSION_MINIMUM=3.5 \
    -DCMAKE_BUILD_TYPE=Release -DDISABLE_SHARED=ON -DDISABLE_TOOL=ON -DDISABLE_CRC=ON
  cmake --build "$work/rapidyenc" --target rapidyenc_static -j 4
  local yenc="$work/rapidyenc/rapidyenc_static/librapidyenc.a"

  (cd "$ROOT" && env CGO_ENABLED=1 GOOS=ios GOARCH="$goarch" CC="$cc" CXX="$cxx" \
    CGO_CFLAGS="-isysroot $sdk_path -target $target -O2" \
    CGO_CXXFLAGS="-isysroot $sdk_path -target $target -O2" \
    CGO_LDFLAGS="-isysroot $sdk_path -target $target $yenc" \
    "$GO" build -trimpath -buildvcs=false -buildmode=c-archive -ldflags="-s -w -buildid=" \
    -o "$work/libgo.a" ./cmd/nuvio-usenet-ios)

  # A C archive leaves libraries named in LDFLAGS to the final link; fold
  # rapidyenc in so the app links one archive per slice.
  libtool -static -o "$work/libnuvio_usenet.a" "$work/libgo.a" "$yenc"
}

build_slice ios-arm64 iphoneos arm64 arm64 "arm64-apple-ios$MIN_IOS"
build_slice simulator-arm64 iphonesimulator arm64 arm64 "arm64-apple-ios$MIN_IOS-simulator"
build_slice simulator-x86_64 iphonesimulator amd64 x86_64 "x86_64-apple-ios$MIN_IOS-simulator"

mkdir -p "$OUT/simulator"
lipo -create "$OUT/simulator-arm64/libnuvio_usenet.a" "$OUT/simulator-x86_64/libnuvio_usenet.a" \
  -output "$OUT/simulator/libnuvio_usenet.a"

xcodebuild -create-xcframework \
  -library "$OUT/ios-arm64/libnuvio_usenet.a" -headers "$OUT/headers" \
  -library "$OUT/simulator/libnuvio_usenet.a" -headers "$OUT/headers" \
  -output "$OUT/NuvioUsenet.xcframework"

(cd "$OUT" && ditto -c -k --keepParent NuvioUsenet.xcframework NuvioUsenet.xcframework.zip \
  && shasum -a 256 NuvioUsenet.xcframework.zip | tee NuvioUsenet.xcframework.zip.sha256)
