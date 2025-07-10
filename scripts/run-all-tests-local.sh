#!/usr/bin/env bash
# scripts/run-all-tests-local.sh
set -euo pipefail

# figure out where we live and where the project root is
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "▶ Running unit tests…"
if ! go test ./...; then
  echo "❌ Unit tests failed"
  exit 1
fi

echo
echo "▶ Running integration tests…"
# forward status code from the integration script
"$SCRIPT_DIR/run-integration-tests-local.sh"
INT_EXIT_CODE=$?
if [ $INT_EXIT_CODE -ne 0 ]; then
  echo "❌ Integration tests failed"
  exit $INT_EXIT_CODE
fi

echo
echo "🎉 All tests passed!"
