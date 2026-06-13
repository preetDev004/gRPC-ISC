#!/usr/bin/env bash
# scripts/gen.sh – run from project root
# Requires: buf (https://buf.build/docs/installation)
set -euo pipefail

echo "→ Generating protobuf / gRPC code..."
buf dep update
buf generate

echo "→ Tidying Go modules..."
go mod tidy

echo "✓ Done. Generated files are in ./gen/"
