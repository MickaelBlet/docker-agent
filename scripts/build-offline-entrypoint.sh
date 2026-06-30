#!/usr/bin/env bash
# Entrypoint of the offline build image. Compiles docker-agent for
# linux/amd64 and windows/amd64 into /out. Runs with no network.
set -euo pipefail

# Built straight from source: version values come from the source tree
# (pkg/version), no git tag/commit injected.
LDFLAGS="-s -w -linkmode=external"

mkdir -p /out

echo ">> Building linux/amd64..."
TARGETPLATFORM=linux/amd64 XX_GO_PREFER_C_COMPILER=zig xx-go build \
  -trimpath -buildvcs=false -tags no_audio -ldflags "$LDFLAGS" \
  -o /out/docker-agent-linux-amd64 .
xx-verify --static /out/docker-agent-linux-amd64

echo ">> Building windows/amd64..."
TARGETPLATFORM=windows/amd64 XX_GO_PREFER_C_COMPILER=zig xx-go build \
  -trimpath -buildvcs=false -tags no_audio -ldflags "$LDFLAGS" \
  -o /out/docker-agent-windows-amd64.exe .

echo ">> Done:"
ls -lh /out
