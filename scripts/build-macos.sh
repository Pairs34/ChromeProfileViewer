#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
PROJECT_DIR=$(dirname "$SCRIPT_DIR")
BUNDLE_DIR="$PROJECT_DIR/build/bin/BrowserProfileViewer.app"
BUILD_CACHE="${TMPDIR:-/tmp}/browser-profile-viewer-go-cache"

mkdir -p "$BUNDLE_DIR/Contents/MacOS" "$BUNDLE_DIR/Contents/Resources" "$BUILD_CACHE"
cp "$PROJECT_DIR/build/darwin/Info.bundle.plist" "$BUNDLE_DIR/Contents/Info.plist"

cd "$PROJECT_DIR"
MACOSX_DEPLOYMENT_TARGET="13.0" \
CGO_LDFLAGS="-framework UniformTypeIdentifiers -mmacosx-version-min=13.0" \
GOCACHE="$BUILD_CACHE" \
go build -buildvcs=false -tags desktop,production \
  -ldflags="-w -s" \
  -o "$BUNDLE_DIR/Contents/MacOS/BrowserProfileViewer"

echo "Built: $BUNDLE_DIR"
