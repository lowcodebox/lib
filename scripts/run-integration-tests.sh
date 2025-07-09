#!/usr/bin/env bash
set -euo pipefail

# locate project root and the compose file
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE_FILE="$PROJECT_ROOT/tests/integration/s3/minio.yml"

if [ ! -f "$COMPOSE_FILE" ]; then
  echo "❌ Could not find compose file: $COMPOSE_FILE"
  exit 1
fi

echo "▶ Starting MinIO for integration tests..."
docker-compose -f "$COMPOSE_FILE" up -d

echo "⏳ Waiting for MinIO to be healthy…"
MAX_RETRIES=10
for i in $(seq 1 $MAX_RETRIES); do
  if curl -fsS "http://localhost:9000/minio/health/live" > /dev/null; then
    echo "✅ MinIO is healthy!"
    break
  fi
  echo "  - retry $i/$MAX_RETRIES…"
  sleep 2
  if [ "$i" -eq "$MAX_RETRIES" ]; then
    echo "❌ MinIO did not become healthy in time"
    docker-compose -f "$COMPOSE_FILE" logs minio
    docker-compose -f "$COMPOSE_FILE" down
    exit 1
  fi
done

echo "▶ Running Go integration tests…"
cd "$PROJECT_ROOT"
go test -timeout 120s -tags=integration ./pkg/s3
TEST_EXIT_CODE=$?

echo "▶ Tearing down MinIO..."
docker-compose -f "$COMPOSE_FILE" down

exit $TEST_EXIT_CODE
