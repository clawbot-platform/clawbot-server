#!/bin/sh
set -eu

if [ -f ./.env ]; then
  set -a
  . ./.env
  set +a
fi

mkdir -p ./.cache/go-build ./.cache/go-mod
GOCACHE="$(pwd)/.cache/go-build" GOMODCACHE="$(pwd)/.cache/go-mod" go run ./cmd/stack-smoke
