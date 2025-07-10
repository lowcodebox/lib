#!/usr/bin/env sh
# scripts/run-unit-tests.sh
set -eu
set -ex

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "▶ Running Go unit tests..."
cd "$PROJECT_ROOT"
go test -v -cover ./...
