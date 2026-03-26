#!/bin/sh
set -eu

if [ -f ./.env ]; then
  set -a
  . ./.env
  set +a
fi

mkdir -p ./.cache/go-build ./.cache/go-mod

# Core-only by default.
# Set VALIDATE_OPTIONAL_STACK=1 when you want smoke checks to include
# optional services from docker-compose.optional.yml.
VALIDATE_OPTIONAL_STACK="${VALIDATE_OPTIONAL_STACK:-0}"

GOCACHE="$(pwd)/.cache/go-build" \
GOMODCACHE="$(pwd)/.cache/go-mod" \
VALIDATE_OPTIONAL_STACK="${VALIDATE_OPTIONAL_STACK}" \
go run ./cmd/stack-smoke
