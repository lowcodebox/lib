#!/usr/bin/env sh
# scripts/run-integration-tests.sh
set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

if [ -z "${MINIO_ENDPOINT:-}" ]; then
  echo "❌ Environment variable MINIO_ENDPOINT is required"
  exit 1
fi

echo "▶ Checking if MinIO is reachable at $MINIO_ENDPOINT.."
if ! curl -fsS "$MINIO_ENDPOINT/minio/health/live" > /dev/null; then
  echo "❌ MinIO is not reachable at $MINIO_ENDPOINT"
  exit 1
fi

echo "✅ MinIO is reachable"

echo "▶ Running Go integration tests..."
cd "$PROJECT_ROOT"
go test -timeout 120s -tags=integration -run '^$' ./...
