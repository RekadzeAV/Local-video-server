#!/usr/bin/env bash
set -euo pipefail
echo "[build-go] static placeholder build"
go build -o ./bin/vigilos-core ./cmd/vigilos-core || true

