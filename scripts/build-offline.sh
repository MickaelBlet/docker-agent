#!/usr/bin/env bash
# Build docker-agent for linux/amd64 and windows/amd64 fully offline,
# straight from the current working directory (mounted at /src).
#
#   1) Once, WITH network:   ./scripts/build-offline.sh prepare
#   2) Anytime, OFFLINE:     ./scripts/build-offline.sh build
#   or both:                 ./scripts/build-offline.sh        (default)
#
# Output binaries land in ./dist :
#   docker-agent-linux-amd64
#   docker-agent-windows-amd64.exe
set -euo pipefail

cd "$(dirname "$0")/.."

IMAGE="docker-agent-build:offline"
DOCKERFILE="Dockerfile.offline"
# Persistent Go build cache (kept out of the image to keep it small; makes
# repeated offline builds fast).
CACHE_VOL="docker-agent-build-cache"

prepare() {
  echo ">> Building offline build image (needs network)..."
  docker build -f "$DOCKERFILE" -t "$IMAGE" .
  echo ">> Image '$IMAGE' ready. Build offline with: $0 build"
}

build() {
  echo ">> Compiling from PWD source OFFLINE (--network=none)..."
  mkdir -p ./dist
  docker run --rm --network=none \
    -v "$CACHE_VOL:/root/.cache/go-build" \
    -v "$PWD:/src" \
    -v "$PWD/dist:/out" \
    "$IMAGE"
  echo ">> Binaries in ./dist :"
  ls -lh ./dist
}

case "${1:-build}" in
  prepare) prepare ;;
  build)   build ;;
  all)     prepare; build ;;
  *) echo "usage: $0 {prepare|build|all}" >&2; exit 1 ;;
esac
