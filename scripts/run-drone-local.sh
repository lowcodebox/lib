#!/usr/bin/env bash
# scripts/run-drone-local.sh
set -euo pipefail
set -ex

# locate project root and the compose file
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE_FILE="$PROJECT_ROOT/tests/integration/s3/minio.yml"

if [ ! -f "$COMPOSE_FILE" ]; then
  echo "❌ Could not find compose file: $COMPOSE_FILE"
  exit 1
fi

echo "▶ Starting MinIO for integration tests..."
docker compose -f "$COMPOSE_FILE" up -d

drone exec --pipeline=lib --trusted --event=push --branch=feature/wpr-test

echo "▶ Tearing down MinIO..."
docker compose -f "$COMPOSE_FILE" down